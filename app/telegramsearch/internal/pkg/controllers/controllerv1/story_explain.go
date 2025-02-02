package controllerv1

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (c *Ctl) ExplainMessage(_ context.Context, senderID int) (string, error) {
	v := c.explainedMessagesCache.Get(senderID)
	if v == nil {
		return c.cfg.NoExplainedAnswer, nil
	}
	return v.Value().ForUser(), nil
}

type ExplainedMessage struct {
	Sources []SourcedMessage
}

func (e ExplainedMessage) ForUser() string {
	result := make([]string, len(e.Sources))
	for i, s := range e.Sources {
		result[i] = s.ForHuman()
	}
	return fmt.Sprintf("%s\n%s\n\n%s",
		"Вот подробные вопросы и ответы пользователей:",
		strings.Join(result, "\n\n"),
		"Если хотите узнать больше, перейдите по ссылке, нажмите правой кнопкой мыши на сообщении и выберите 'Просмотреть ответы'. Вы можете задать уточняющий вопрос когда перейдете по ссылке, если вам мало информации. Помните, что у вас может не быть доступа до указанного чата. Сперва придется в него вступить.",
	)
}

type SourcedMessage struct {
	URL                 string
	MostRecentMessageAt time.Time
	// A couple of letters from the first message, that give enough context.
	FirstLetters string
}

func (m SourcedMessage) ForHuman() string {
	return fmt.Sprintf("%s: %s\n%s", m.MostRecentMessageAt.Format(time.DateOnly), m.URL, m.FirstLetters)
}
