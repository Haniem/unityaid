package shop

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInsufficientCoins = errors.New("insufficient coins")
var ErrOutOfStock = errors.New("product out of stock")
var ErrNotFound = errors.New("not found")
var ErrInvalidStatus = errors.New("invalid status")
var ErrInvalidStatusTransition = errors.New("invalid status transition")
var ErrMixedOrganizations = errors.New("cart contains products from different organizations")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Wallet(ctx context.Context, userID string) (Wallet, error) {
	var wallet Wallet
	wallet.UserID = userID
	if err := r.db.QueryRow(ctx, `SELECT COALESCE(coin_balance, 0) FROM volunteer_profiles WHERE user_id = $1`, userID).Scan(&wallet.Balance); err != nil {
		return Wallet{}, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id::text, user_id::text, amount, type, description, created_at
		FROM coin_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 20
	`, userID)
	if err != nil {
		return Wallet{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var item CoinTransaction
		if err := rows.Scan(&item.ID, &item.UserID, &item.Amount, &item.Type, &item.Description, &item.CreatedAt); err != nil {
			return Wallet{}, err
		}
		wallet.Transactions = append(wallet.Transactions, item)
	}
	return wallet, rows.Err()
}

func (r *Repository) Transfer(ctx context.Context, fromUserID string, req TransferRequest) (Wallet, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Wallet{}, err
	}
	defer tx.Rollback(ctx)

	var balance int
	if err := tx.QueryRow(ctx, `SELECT coin_balance FROM volunteer_profiles WHERE user_id = $1 FOR UPDATE`, fromUserID).Scan(&balance); err != nil {
		return Wallet{}, err
	}
	if balance < req.Amount {
		return Wallet{}, ErrInsufficientCoins
	}
	if _, err := tx.Exec(ctx, `INSERT INTO volunteer_profiles (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`, req.RecipientID); err != nil {
		return Wallet{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE volunteer_profiles SET coin_balance = coin_balance - $2 WHERE user_id = $1`, fromUserID, req.Amount); err != nil {
		return Wallet{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE volunteer_profiles SET coin_balance = coin_balance + $2 WHERE user_id = $1`, req.RecipientID, req.Amount); err != nil {
		return Wallet{}, err
	}
	description := req.Comment
	if description == "" {
		description = "Перевод монет"
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO coin_transactions (user_id, amount, type, description, related_user_id)
		VALUES ($1, $2, 'transfer_out', $4, $5), ($5, $3, 'transfer_in', $4, $1)
	`, fromUserID, -req.Amount, req.Amount, description, req.RecipientID); err != nil {
		return Wallet{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Wallet{}, err
	}
	return r.Wallet(ctx, fromUserID)
}

func (r *Repository) Products(ctx context.Context, organizationIDs []string) ([]Product, error) {
	if organizationIDs != nil && len(organizationIDs) == 0 {
		return []Product{}, nil
	}
	showAll := organizationIDs == nil
	organizationIDsArg := organizationIDs
	if organizationIDsArg == nil {
		organizationIDsArg = []string{}
	}
	rows, err := r.db.Query(ctx, `
		SELECT p.id::text, p.organization_id::text, o.name, p.name, p.description, p.price, p.stock, p.image_url, p.is_active, p.created_at, p.updated_at
		FROM shop_products p
		LEFT JOIN organizations o ON o.id = p.organization_id
		WHERE p.is_active = true
			AND p.organization_id IS NOT NULL
			AND ($1::bool OR p.organization_id::text = ANY($2::text[]))
		ORDER BY p.price, p.name
	`, showAll, organizationIDsArg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Product{}
	for rows.Next() {
		item, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateProduct(ctx context.Context, req CreateProductRequest) (Product, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return scanProduct(r.db.QueryRow(ctx, `
		INSERT INTO shop_products (organization_id, name, description, price, stock, image_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, organization_id::text, NULL::text, name, description, price, stock, image_url, is_active, created_at, updated_at
	`, req.OrganizationID, req.Name, req.Description, req.Price, req.Stock, req.ImageURL, isActive))
}

func (r *Repository) CreateOrder(ctx context.Context, userID string, organizationIDs []string, req CreateOrderRequest) (Order, error) {
	if organizationIDs != nil && len(organizationIDs) == 0 {
		return Order{}, ErrNotFound
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx)

	var balance int
	if err := tx.QueryRow(ctx, `SELECT coin_balance FROM volunteer_profiles WHERE user_id = $1 FOR UPDATE`, userID).Scan(&balance); err != nil {
		return Order{}, err
	}
	total := 0
	showAll := organizationIDs == nil
	organizationIDsArg := organizationIDs
	if organizationIDsArg == nil {
		organizationIDsArg = []string{}
	}
	type productSnapshot struct {
		id, name string
		orgID    *string
		price    int
		quantity int
	}
	products := []productSnapshot{}
	for _, requested := range req.Items {
		var product productSnapshot
		var orgID *string
		if err := tx.QueryRow(ctx, `
			SELECT id::text, organization_id::text, name, price, stock
			FROM shop_products
			WHERE id = $1
				AND is_active = true
				AND organization_id IS NOT NULL
				AND ($2::bool OR organization_id::text = ANY($3::text[]))
			FOR UPDATE
		`, requested.ProductID, showAll, organizationIDsArg).Scan(&product.id, &orgID, &product.name, &product.price, &product.quantity); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return Order{}, ErrNotFound
			}
			return Order{}, err
		}
		if product.quantity < requested.Quantity {
			return Order{}, ErrOutOfStock
		}
		product.orgID = orgID
		product.quantity = requested.Quantity
		if len(products) > 0 && (products[0].orgID == nil || product.orgID == nil || *products[0].orgID != *product.orgID) {
			return Order{}, ErrMixedOrganizations
		}
		total += product.price * requested.Quantity
		products = append(products, product)
	}
	if balance < total {
		return Order{}, ErrInsufficientCoins
	}

	var organizationID *string
	if len(products) > 0 {
		organizationID = products[0].orgID
	}
	var orderID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO shop_orders (user_id, organization_id, total, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text
	`, userID, organizationID, total, req.Comment).Scan(&orderID); err != nil {
		return Order{}, err
	}
	for _, product := range products {
		if _, err := tx.Exec(ctx, `
			INSERT INTO shop_order_items (order_id, product_id, product_name, price, quantity)
			VALUES ($1, $2, $3, $4, $5)
		`, orderID, product.id, product.name, product.price, product.quantity); err != nil {
			return Order{}, err
		}
		if _, err := tx.Exec(ctx, `UPDATE shop_products SET stock = stock - $2, updated_at = now() WHERE id = $1`, product.id, product.quantity); err != nil {
			return Order{}, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE volunteer_profiles SET coin_balance = coin_balance - $2, updated_at = now() WHERE user_id = $1`, userID, total); err != nil {
		return Order{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO coin_transactions (user_id, amount, type, description, source_type, source_id)
		VALUES ($1, $2, 'purchase', 'Покупка в корпоративном магазине', 'shop_order', $3)
	`, userID, -total, orderID); err != nil {
		return Order{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, err
	}
	return r.Order(ctx, orderID)
}

func (r *Repository) Orders(ctx context.Context, userID string, all bool, organizationIDs []string) ([]Order, error) {
	if all && organizationIDs != nil && len(organizationIDs) == 0 {
		return []Order{}, nil
	}
	sql := `
		SELECT o.id::text, o.user_id::text, concat_ws(' ', u.last_name, u.first_name), o.organization_id::text, org.name, o.status, o.total, o.comment, o.created_at, o.updated_at
		FROM shop_orders o
		JOIN users u ON u.id = o.user_id
		LEFT JOIN organizations org ON org.id = o.organization_id
	`
	args := []any{}
	if !all {
		sql += ` WHERE o.user_id = $1`
		args = append(args, userID)
	} else if organizationIDs != nil {
		sql += ` WHERE o.organization_id::text = ANY($1::text[])`
		args = append(args, organizationIDs)
	}
	sql += ` ORDER BY o.created_at DESC`
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Order{}
	for rows.Next() {
		item, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		item.Items, err = r.orderItems(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Order(ctx context.Context, id string) (Order, error) {
	item, err := scanOrder(r.db.QueryRow(ctx, `
		SELECT o.id::text, o.user_id::text, concat_ws(' ', u.last_name, u.first_name), o.organization_id::text, org.name, o.status, o.total, o.comment, o.created_at, o.updated_at
		FROM shop_orders o
		JOIN users u ON u.id = o.user_id
		LEFT JOIN organizations org ON org.id = o.organization_id
		WHERE o.id = $1
	`, id))
	if err != nil {
		return Order{}, err
	}
	item.Items, err = r.orderItems(ctx, item.ID)
	return item, err
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, id string, status string, managerID string) (Order, error) {
	if !isAllowedOrderStatus(status) {
		return Order{}, ErrInvalidStatus
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx)

	var userID, currentStatus string
	var total int
	if err := tx.QueryRow(ctx, `
		SELECT user_id::text, status, total
		FROM shop_orders
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&userID, &currentStatus, &total); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrNotFound
		}
		return Order{}, err
	}
	if currentStatus == "cancelled" && status != "cancelled" {
		return Order{}, ErrInvalidStatusTransition
	}
	if status == "cancelled" && currentStatus != "cancelled" {
		if _, err := tx.Exec(ctx, `
			UPDATE volunteer_profiles
			SET coin_balance = coin_balance + $2, updated_at = now()
			WHERE user_id = $1
		`, userID, total); err != nil {
			return Order{}, err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE shop_products product
			SET stock = product.stock + item.quantity, updated_at = now()
			FROM shop_order_items item
			WHERE item.order_id = $1 AND item.product_id = product.id
		`, id); err != nil {
			return Order{}, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO coin_transactions (user_id, amount, type, description, source_type, source_id)
			VALUES ($1, $2, 'refund', 'Возврат за отмененный заказ', 'shop_order', $3)
		`, userID, total, id); err != nil {
			return Order{}, err
		}
	}

	tag, err := tx.Exec(ctx, `
		UPDATE shop_orders
		SET status = $2, processed_by = $3, processed_at = now(), updated_at = now()
		WHERE id = $1
	`, id, status, managerID)
	if err != nil {
		return Order{}, err
	}
	if tag.RowsAffected() == 0 {
		return Order{}, ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, err
	}
	return r.Order(ctx, id)
}

func isAllowedOrderStatus(status string) bool {
	switch status {
	case "pending", "processing", "completed", "cancelled":
		return true
	default:
		return false
	}
}

func (r *Repository) orderItems(ctx context.Context, orderID string) ([]OrderItem, error) {
	rows, err := r.db.Query(ctx, `SELECT id::text, product_id::text, product_name, price, quantity FROM shop_order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []OrderItem{}
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.ProductName, &item.Price, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(row scanner) (Product, error) {
	var item Product
	err := row.Scan(&item.ID, &item.OrganizationID, &item.OrganizationName, &item.Name, &item.Description, &item.Price, &item.Stock, &item.ImageURL, &item.IsActive, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func scanOrder(row scanner) (Order, error) {
	var item Order
	err := row.Scan(&item.ID, &item.UserID, &item.UserName, &item.OrganizationID, &item.OrganizationName, &item.Status, &item.Total, &item.Comment, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
