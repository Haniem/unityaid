<script setup lang="ts">
import { computed, ref } from 'vue'
import { Award, BarChart3, BookOpen, CalendarDays, CheckCircle2, ChevronDown, HelpCircle, Newspaper, ShieldCheck, ShoppingCart, UsersRound } from 'lucide-vue-next'

type HelpItem = {
  id: string
  title: string
  icon: object
  text: string
  steps: string[]
  result: string
}

const sections: { id: string; label: string; items: HelpItem[] }[] = [
  {
    id: 'start',
    label: 'С чего начать',
    items: [
      {
        id: 'tour',
        title: 'Быстрый тур по платформе',
        icon: BookOpen,
        text: 'Пульс объединяет мероприятия, задачи, волонтеров, новости, достижения, магазин и аналитику.',
        steps: [
          'Откройте главную страницу и проверьте ближайшие события, активные задачи и уведомления.',
          'Перейдите в календарь, откройте событие и покажите заявки, смены, посещаемость и завершение мероприятия.',
          'Откройте задачи: назначение исполнителя, комментарии и подтверждение выполнения.',
          'Завершите сценарий профилем волонтера, достижениями, монетами и магазином.'
        ],
        result: 'После тура видно полный цикл: планирование, участие, учет результата и мотивация.'
      },
      {
        id: 'roles',
        title: 'Роли пользователей',
        icon: ShieldCheck,
        text: 'Для защиты подготовлены четыре пользователя: главный админ, админ организации, менеджер и волонтер.',
        steps: [
          'admin@puls.test видит системные настройки, аналитику и все организации.',
          'admin1@puls.test управляет организацией, сотрудниками, новостями и заказами магазина.',
          'admin2@puls.test координирует события, задачи, заявки, часы и заказы.',
          'admin3@puls.test подает заявки, выполняет задачи, получает достижения и тратит монеты.'
        ],
        result: 'Разделение ролей показывает, что интерфейс адаптируется под обязанности пользователя.'
      }
    ]
  },
  {
    id: 'modules',
    label: 'Основные модули',
    items: [
      {
        id: 'events',
        title: 'Мероприятия',
        icon: CalendarDays,
        text: 'Календарь хранит события, смены, заявки, посещаемость и начисление часов.',
        steps: [
          'Создайте или откройте опубликованное мероприятие.',
          'Покажите заявку волонтера и изменение ее статуса.',
          'Отметьте посещаемость и укажите часы.',
          'Завершите мероприятие, чтобы часы ушли в профиль и аналитику.'
        ],
        result: 'Менеджер видит операционный экран, волонтер видит свое участие и подтвержденные часы.'
      },
      {
        id: 'tasks',
        title: 'Задачи',
        icon: CheckCircle2,
        text: 'Задачи помогают распределять работу внутри события или организации.',
        steps: [
          'Создайте задачу из плюсика в хедере или на странице задач.',
          'Назначьте исполнителя через селект с поиском пользователя.',
          'Добавьте комментарий, вложение в редактировании и учет времени.',
          'Подтвердите выполнение, чтобы результат попал в историю.'
        ],
        result: 'На защите удобно показать прозрачный путь задачи от постановки до подтверждения.'
      },
      {
        id: 'volunteers',
        title: 'Волонтеры и профили',
        icon: UsersRound,
        text: 'Раздел волонтеров показывает карточки участников, навыки, статус, опыт и переход в профиль.',
        steps: [
          'Найдите волонтера через поиск или фильтр по навыку.',
          'Откройте чужой профиль из карточки волонтера.',
          'Покажите вкладки: личные данные, задачи, работа, мероприятия, достижения и документы.',
          'Для своего профиля дополнительно покажите редактирование данных и смену пароля.'
        ],
        result: 'Профиль становится центром истории участника, а не только настройкой аккаунта.'
      },
      {
        id: 'news',
        title: 'Новости и база знаний',
        icon: Newspaper,
        text: 'Новости закрывают публичные анонсы и отчеты, база знаний — инструкции и внутренние материалы.',
        steps: [
          'Создайте новость через плюсик в хедере или кнопку на странице.',
          'Заполните категорию, обложку, статус публикации и текст.',
          'Покажите список новостей с фильтрами и карточку новости.',
          'В базе знаний откройте инструкцию для команды.'
        ],
        result: 'Коммуникационный блок показывает, как организация информирует волонтеров.'
      },
      {
        id: 'shop',
        title: 'Магазин и монеты',
        icon: ShoppingCart,
        text: 'Монеты начисляются за достижения и тратятся в корпоративном магазине.',
        steps: [
          'Откройте магазин и покажите баланс пользователя.',
          'Добавьте товар в корзину и оплатите монетами.',
          'Покажите проверку баланса и историю транзакций.',
          'Под менеджером обработайте заказ: взять в работу, выдать или отменить.'
        ],
        result: 'Магазин завершает мотивационный цикл: активность превращается в внутренние награды.'
      }
    ]
  },
  {
    id: 'control',
    label: 'Контроль и отчеты',
    items: [
      {
        id: 'analytics',
        title: 'Аналитика',
        icon: BarChart3,
        text: 'Аналитика собирает показатели по событиям, задачам, волонтерам, часам и геймификации.',
        steps: [
          'Откройте сводку и сравните ключевые показатели.',
          'Перейдите в отчет по мероприятиям или задачам.',
          'Покажите выгрузку отчета для руководителя.',
          'Сопоставьте цифры с действиями, выполненными в сценарии.'
        ],
        result: 'Руководитель получает экран принятия решений без ручного сбора данных.'
      },
      {
        id: 'achievements',
        title: 'Достижения',
        icon: Award,
        text: 'Достижения фиксируют активность волонтера и автоматически начисляют монеты.',
        steps: [
          'Откройте страницу достижений и лидерборд.',
          'Покажите полученные достижения в профиле волонтера.',
          'Объясните, что монеты начисляются при появлении нового достижения.',
          'Перейдите в магазин и покажите пополнение баланса.'
        ],
        result: 'Геймификация связывает часы, задачи, достижения и магазин в единый сценарий.'
      },
      {
        id: 'faq',
        title: 'Частые вопросы',
        icon: HelpCircle,
        text: 'FAQ используется как встроенный справочник для демонстрации и обучения сотрудников.',
        steps: [
          'Выберите тему слева.',
          'Следуйте шагам выбранного гайда.',
          'Переходите между модулями без выхода из приложения.',
          'Используйте этот раздел как сценарий показа на защите.'
        ],
        result: 'Вместо видео здесь собраны короткие прикладные инструкции по основным модулям.'
      }
    ]
  }
]

const activeSectionId = ref(sections[0].id)
const activeItemId = ref(sections[0].items[0].id)
const isHelpNavOpen = ref(false)

const activeItem = computed(() => sections.flatMap((section) => section.items).find((item) => item.id === activeItemId.value) ?? sections[0].items[0])
const activeSection = computed(() => sections.find((section) => section.id === activeSectionId.value) ?? sections[0])

function selectItem(sectionId: string, itemId: string) {
  activeSectionId.value = sectionId
  activeItemId.value = itemId
  isHelpNavOpen.value = false
}
</script>

<template>
  <section class="page-section faq-page">
    <div class="help-shell">
      <aside class="help-nav" aria-label="Разделы помощи">
        <div class="help-nav-title">
          <BookOpen :size="20" />
          <strong>Помощь</strong>
        </div>
        <button class="help-nav-toggle" :class="{ open: isHelpNavOpen }" type="button" @click="isHelpNavOpen = !isHelpNavOpen">
          <span>
            <small>Разделы помощи</small>
            <strong>{{ activeItem.title }}</strong>
          </span>
          <ChevronDown :size="18" />
        </button>
        <div class="help-nav-list" :class="{ open: isHelpNavOpen }">
          <div v-for="section in sections" :key="section.id" class="help-nav-group">
            <p>{{ section.label }}</p>
            <button v-for="item in section.items" :key="item.id" :class="{ active: activeItemId === item.id }" type="button" @click="selectItem(section.id, item.id)">
              <component :is="item.icon" :size="16" />
              <span>{{ item.title }}</span>
            </button>
          </div>
        </div>
      </aside>

      <article class="help-content">
        <p class="eyebrow">{{ activeSection.label }}</p>
        <h1>{{ activeItem.title }}</h1>
        <p>{{ activeItem.text }}</p>

        <section class="help-guide-panel">
          <h2>Гайд по использованию</h2>
          <ol>
            <li v-for="step in activeItem.steps" :key="step">{{ step }}</li>
          </ol>
        </section>

        <section class="help-result-panel">
          <h2>Где искать результат</h2>
          <p>{{ activeItem.result }}</p>
        </section>
      </article>
    </div>
  </section>
</template>
