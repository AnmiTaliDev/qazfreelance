// SPDX-FileCopyrightText: AnmiTaliDev <anmitalidev@nuros.org>
// SPDX-License-Identifier: AGPL-3.0-only

package i18n

import "fmt"

const DefaultLang = "en"

var strings = map[string]map[string]string{
	"choose_language": {
		"kk": "Тілді таңдаңыз:",
		"ru": "Пожалуйста, выберите язык:",
		"en": "Please choose your language:",
	},
	"btn_kk": {
		"kk": "Казакша",
		"ru": "Казахский",
		"en": "Kazakh",
	},
	"btn_ru": {
		"kk": "Орысша",
		"ru": "Русский",
		"en": "Russian",
	},
	"btn_en": {
		"kk": "Ағылшынша",
		"ru": "Английский",
		"en": "English",
	},
	"main_menu": {
		"kk": "Басты мәзір. Әрекетті таңдаңыз:",
		"ru": "Главное меню. Выберите действие:",
		"en": "Main menu. Choose an action:",
	},
	"btn_post_order": {
		"kk": "Тапсырыс беру",
		"ru": "Разместить заказ",
		"en": "Post an order",
	},
	"btn_advertise": {
		"kk": "Өзіңді жарнамала",
		"ru": "Предложить услуги",
		"en": "Advertise yourself",
	},
	"btn_moderator_menu": {
		"kk": "Модератор мәзірі",
		"ru": "Меню модератора",
		"en": "Moderator menu",
	},
	"enter_title": {
		"kk": "Тақырыпты енгізіңіз:",
		"ru": "Введите заголовок:",
		"en": "Enter a title for your submission:",
	},
	"enter_description": {
		"kk": "Сипаттаманы енгізіңіз:",
		"ru": "Введите описание:",
		"en": "Enter a description:",
	},
	"enter_contact": {
		"kk": "Байланыс ақпаратын енгізіңіз (телефон, Telegram, электрондық пошта):",
		"ru": "Введите контактную информацию (телефон, Telegram, email):",
		"en": "Enter your contact information (phone, Telegram, email):",
	},
	"submission_received": {
		"kk": "Сіздің өтінімніз қабылданды және модерацияға жіберілді. Шешім туралы хабарлаймыз.",
		"ru": "Ваша заявка принята и отправлена на модерацию. Мы уведомим вас о решении.",
		"en": "Your submission has been received and sent for moderation. We will notify you of the decision.",
	},
	"submission_approved": {
		"kk": "Сіздің өтінімніз бекітілді және арнаға жарияланды.",
		"ru": "Ваша заявка одобрена и опубликована в канале.",
		"en": "Your submission has been approved and published.",
	},
	"submission_rejected": {
		"kk": "Сіздің өтінімніз модераторлар тарапынан қабылданбады.",
		"ru": "Ваша заявка отклонена модераторами.",
		"en": "Your submission has been rejected by moderators.",
	},
	"btn_approve": {
		"kk": "Бекіту",
		"ru": "Одобрить",
		"en": "Approve",
	},
	"btn_reject": {
		"kk": "Бас тарту",
		"ru": "Отклонить",
		"en": "Reject",
	},
	"btn_open_list": {
		"kk": "Тізімді ашу",
		"ru": "Открыть список",
		"en": "Open list",
	},
	"btn_conflicts": {
		"kk": "Қарама-қайшылықтар",
		"ru": "Конфликты",
		"en": "Conflicts",
	},
	"btn_toggle_mode": {
		"kk": "Режимді ауыстыру",
		"ru": "Переключить режим",
		"en": "Switch mode",
	},
	"btn_resolve_approve": {
		"kk": "Шешу: Бекіту",
		"ru": "Разрешить: Одобрить",
		"en": "Resolve: Approve",
	},
	"btn_resolve_reject": {
		"kk": "Шешу: Бас тарту",
		"ru": "Разрешить: Отклонить",
		"en": "Resolve: Reject",
	},
	"moderator_menu": {
		"kk": "Модератор мәзірі:",
		"ru": "Меню модератора:",
		"en": "Moderator menu:",
	},
	"mode_current_stream": {
		"kk": "Ағымдағы режим: ағын (өтінімдер дереу келеді)",
		"ru": "Текущий режим: поток (заявки поступают мгновенно)",
		"en": "Current mode: stream (submissions arrive instantly)",
	},
	"mode_current_list": {
		"kk": "Ағымдағы режим: тізім (өтінімдер кезекке қосылады)",
		"ru": "Текущий режим: список (заявки накапливаются в очереди)",
		"en": "Current mode: list (submissions are queued)",
	},
	"mode_switched_stream": {
		"kk": "Режим ауысты: ағын",
		"ru": "Режим переключён: поток",
		"en": "Mode switched to: stream",
	},
	"mode_switched_list": {
		"kk": "Режим ауысты: тізім",
		"ru": "Режим переключён: список",
		"en": "Mode switched to: list",
	},
	"pending_notify": {
		"kk": "Жаңа өтінімдер тексеруді күтеді. Оларды көру үшін 'Тізімді ашу' батырмасын басыңыз.",
		"ru": "Новые заявки ожидают проверки. Нажмите 'Открыть список', чтобы просмотреть их.",
		"en": "New submissions are waiting for your review. Press 'Open list' to view them.",
	},
	"no_pending": {
		"kk": "Күтілетін өтінімдер жоқ.",
		"ru": "Нет ожидающих заявок.",
		"en": "No pending submissions.",
	},
	"no_conflicts": {
		"kk": "Қарама-қайшылықтар жоқ.",
		"ru": "Нет конфликтов.",
		"en": "No conflicts.",
	},
	"submission_info": {
		"kk": "Өтінім #%d\nТүрі: %s\nТақырып: %s\nСипаттама: %s\nБайланыс: %s\nКүй: %s",
		"ru": "Заявка #%d\nТип: %s\nЗаголовок: %s\nОписание: %s\nКонтакт: %s\nСтатус: %s",
		"en": "Submission #%d\nType: %s\nTitle: %s\nDescription: %s\nContact: %s\nStatus: %s",
	},
	"submission_type_order": {
		"kk": "Тапсырыс",
		"ru": "Заказ",
		"en": "Order",
	},
	"submission_type_resume": {
		"kk": "Жарнама",
		"ru": "Реклама",
		"en": "Advertisement",
	},
	"conflict_notify": {
		"kk": "Өтінім #%d бойынша қарама-қайшы шешімдер бар. Оны шешіңіз.",
		"ru": "По заявке #%d есть противоречивые решения. Пожалуйста, разрешите конфликт.",
		"en": "Submission #%d has conflicting decisions. Please resolve it.",
	},
	"already_decided": {
		"kk": "Сіз бұл өтінім бойынша шешім қабылдадыңыз.",
		"ru": "Вы уже приняли решение по этой заявке.",
		"en": "You have already made a decision on this submission.",
	},
	"unknown_command": {
		"kk": "Белгісіз команда. Мәзір батырмаларын пайдаланыңыз.",
		"ru": "Неизвестная команда. Используйте кнопки меню.",
		"en": "Unknown command. Use the menu buttons.",
	},
	"cancel": {
		"kk": "Болдырылмады. Басты мәзірге оралу.",
		"ru": "Отменено. Возврат в главное меню.",
		"en": "Cancelled. Returning to main menu.",
	},
	"btn_cancel": {
		"kk": "Болдырмау",
		"ru": "Отмена",
		"en": "Cancel",
	},
	"btn_next": {
		"kk": "Келесі",
		"ru": "Следующий",
		"en": "Next",
	},
	"submission_new": {
		"kk": "Жаңа өтінім келіп түсті:",
		"ru": "Поступила новая заявка:",
		"en": "New submission received:",
	},
	"list_end": {
		"kk": "Тізімнің соңы.",
		"ru": "Конец списка.",
		"en": "End of list.",
	},
	"decision_recorded": {
		"kk": "Шешіміңіз тіркелді.",
		"ru": "Ваше решение зафиксировано.",
		"en": "Your decision has been recorded.",
	},
	"btn_guided": {
		"kk": "Қадамдық",
		"ru": "По шагам",
		"en": "Step by step",
	},
	"btn_free_form": {
		"kk": "Еркін мәтін",
		"ru": "Свободная форма",
		"en": "Free form",
	},
	"choose_submission_mode": {
		"kk": "Өтінімді қалай толтырғыңыз келеді?",
		"ru": "Как вы хотите заполнить заявку?",
		"en": "How would you like to fill in your submission?",
	},
	"enter_free_form": {
		"kk": "Хабарламаңызды жіберіңіз. Фото тіркеуге болады, мәтін немесе тақырып сипаттама ретінде қолданылады.",
		"ru": "Отправьте ваше сообщение. Можно прикрепить фото — текст или подпись будет использована как описание.",
		"en": "Send your message. You can attach a photo — the text or caption will be used as the description.",
	},
	"photo_not_in_title": {
		"kk": "Тақырып үшін мәтін жіберіңіз.",
		"ru": "Для заголовка отправьте текст.",
		"en": "Please send text for the title.",
	},
	"photo_not_in_contact": {
		"kk": "Байланыс ақпараты үшін мәтін жіберіңіз.",
		"ru": "Для контактной информации отправьте текст.",
		"en": "Please send text for your contact information.",
	},
	"contact_label": {
		"kk": "Байланыс: %s",
		"ru": "Контакт: %s",
		"en": "Contact: %s",
	},
	"btn_back": {
		"kk": "Артқа",
		"ru": "Назад",
		"en": "Back",
	},
	"btn_contact": {
		"kk": "Байланысу",
		"ru": "Написать",
		"en": "Contact",
	},
	"btn_my_submissions": {
		"kk": "Менің хабарландыруларым",
		"ru": "Мои объявления",
		"en": "My submissions",
	},
	"my_submissions_header": {
		"kk": "Сіздің өтінімдеріңіз:",
		"ru": "Ваши заявки:",
		"en": "Your submissions:",
	},
	"no_submissions": {
		"kk": "Сізде әлі өтінімдер жоқ.",
		"ru": "У вас пока нет заявок.",
		"en": "You have no submissions yet.",
	},
	"btn_withdraw": {
		"kk": "Алып тастау",
		"ru": "Снять",
		"en": "Withdraw",
	},
	"submission_withdrawn": {
		"kk": "Өтінім алынды.",
		"ru": "Объявление снято.",
		"en": "Submission withdrawn.",
	},
	"cannot_withdraw": {
		"kk": "Бұл өтінімді алу мүмкін емес.",
		"ru": "Это объявление нельзя снять.",
		"en": "This submission cannot be withdrawn.",
	},
	"status_pending": {
		"kk": "Модерацияда",
		"ru": "На модерации",
		"en": "Pending",
	},
	"status_approved": {
		"kk": "Жарияланды",
		"ru": "Опубликовано",
		"en": "Published",
	},
	"status_rejected": {
		"kk": "Қабылданбады",
		"ru": "Отклонено",
		"en": "Rejected",
	},
	"status_conflict": {
		"kk": "Қарама-қайшылық",
		"ru": "Конфликт",
		"en": "Conflict",
	},
	"status_withdrawn": {
		"kk": "Алынды",
		"ru": "Снято",
		"en": "Withdrawn",
	},
	"channel_withdrawn_notice": {
		"kk": "Хабарландыру авторымен алынды.",
		"ru": "Объявление снято автором.",
		"en": "This post has been withdrawn by the author.",
	},
	"user_submission_card": {
		"kk": "#%d — %s\n%s",
		"ru": "#%d — %s\n%s",
		"en": "#%d — %s\n%s",
	},
}

func T(lang, key string) string {
	translations, ok := strings[key]
	if !ok {
		return key
	}
	if val, ok := translations[lang]; ok {
		return val
	}
	if val, ok := translations[DefaultLang]; ok {
		return val
	}
	return key
}

func Tf(lang, key string, args ...any) string {
	return fmt.Sprintf(T(lang, key), args...)
}
