INSERT INTO knowledge_categories (name, slug, description)
VALUES
  ('Старт волонтера', 'start-volunteer', 'Материалы для знакомства с платформой и первыми шагами.'),
  ('Мероприятия', 'events', 'Инструкции по участию и координации мероприятий.'),
  ('Задачи', 'tasks', 'Справка по задачам, срокам и отчетности.')
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    description = EXCLUDED.description;

INSERT INTO knowledge_articles (id, category_id, title, slug, summary, content_html, status, author_id, published_at)
VALUES
  (
    '81818181-8181-8181-8181-818181818181',
    (SELECT id FROM knowledge_categories WHERE slug = 'start-volunteer'),
    'Как начать работу в Пульсе',
    'start-guide',
    'Короткий маршрут для первого входа: профиль, мероприятия, задачи и часы.',
    '<img src="https://images.unsplash.com/photo-1559027615-cd4628902d4a?auto=format&fit=crop&w=1200&q=80" alt="Команда волонтеров" /><h2>Первый вход</h2><p>Проверьте профиль, контактные данные и интересы. После этого откройте календарь мероприятий и выберите ближайшую смену.</p><h2>Что показать на защите</h2><p>Зайдите под волонтером, подайте заявку на мероприятие, выполните задачу и отправьте запись времени на проверку.</p>',
    'published',
    (SELECT id FROM users WHERE email = 'admin1@puls.test'),
    now() - interval '5 days'
  ),
  (
    '82828282-8282-8282-8282-828282828282',
    (SELECT id FROM knowledge_categories WHERE slug = 'events'),
    'Проведение мероприятия',
    'event-management-guide',
    'Как создать мероприятие, подтвердить заявки, отметить посещаемость и начислить часы.',
    '<img src="https://images.unsplash.com/photo-1517048676732-d65bc937f952?auto=format&fit=crop&w=1200&q=80" alt="Планирование мероприятия" /><h2>Подготовка</h2><p>Создайте мероприятие, заполните место, лимит участников и описание. Для повторяющихся событий используйте переключатель в форме создания.</p><h2>В день события</h2><p>Координатор подтверждает заявки, отмечает посещаемость и после завершения начисляет часы участникам.</p>',
    'published',
    (SELECT id FROM users WHERE email = 'admin2@puls.test'),
    now() - interval '4 days'
  ),
  (
    '83838383-8383-8383-8383-838383838383',
    (SELECT id FROM knowledge_categories WHERE slug = 'tasks'),
    'Работа с задачами',
    'task-workflow-guide',
    'Жизненный цикл задачи: создание, назначение, выполнение, проверка и история времени.',
    '<img src="https://images.unsplash.com/photo-1454165804606-c3d57bc86b40?auto=format&fit=crop&w=1200&q=80" alt="Рабочая доска задач" /><h2>Назначение</h2><p>Менеджер создает задачу и выбирает исполнителя через поиск пользователя. Исполнитель видит задачу в списке и подтверждает выполнение.</p><h2>Контроль</h2><p>Комментарии и учет времени остаются на детальной странице, а вложения добавляются при редактировании задачи.</p>',
    'published',
    (SELECT id FROM users WHERE email = 'admin2@puls.test'),
    now() - interval '3 days'
  )
ON CONFLICT (slug) DO UPDATE
SET category_id = EXCLUDED.category_id,
    title = EXCLUDED.title,
    summary = EXCLUDED.summary,
    content_html = EXCLUDED.content_html,
    status = EXCLUDED.status,
    author_id = EXCLUDED.author_id,
    published_at = EXCLUDED.published_at;
