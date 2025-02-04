package controllerv1

import (
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
)

type Config struct {
	HelpText           string `yaml:"help_text"`
	NewsText           string `yaml:"news_text"`
	NoExplainedAnswer  string `yaml:"no_explained_answer"`
	NoResultsAnswer    string
	StaleResponsesText string
	FreshResponsesText string
	StaleThreshold     encodingtooling.Duration
}

func DefaultConfig() Config {
	return Config{
		NoExplainedAnswer: "К сожалению, у меня нет информации об этом сообщении. Возможно оно было задано слишком давно и я о нем забыл.",
		HelpText: `Добрый день! Это - бот, который поможет вам найти ответы на ваши вопросы.

Чтобы задать вопрос, просто напишите его в чат. Он постарается найти наиболее подходящий ответ на ваш вопрос.
Бот не выдумывает от себя, все его данные собраны от живых людей.

Я, разработчик, буду очень признателен, если вы поделитесь своими впечатлениями о боте и порекомендуете его своим друзьям, если он вам полезен.

Вот тут можно связаться с разработчиком: https://substack.com/home/post/p-148053843
`,
		NewsText: `Новости:

Мы открыли чат, в котором все вопросы можно задавать и боту, и людям. Присоединяйтесь: https://t.me/+8trW_-0GEFI1NTE0
Мы планируем добавить в бота функцию показа ссылок на чаты, где было найдена эта информация.
Это позволит вам быстро найти чаты, где обсуждаются интересные вам темы.
Если вам есть что прокомментировать или добавить - пишите в чате https://t.me/+8trW_-0GEFI1NTE0 или в https://substack.com/home/post/p-148053843
`,
		StaleThreshold:     encodingtooling.Duration{time.Hour * 24 * 365 * 2},
		StaleResponsesText: "В ответе не использовано информации свежее чем от %s",
		FreshResponsesText: "Обсуждений: %d",
		NoResultsAnswer: "К сожалению, у меня недостаточно информации чтобы ответить на данный вопрос." +
			" Но я учусь и, вероятно, смогу ответить на него позже.",
	}
}
