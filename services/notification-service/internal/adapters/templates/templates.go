package templates

import (
	"bytes"
	"html/template"
	"time"

	"todoapp/services/notification-service/internal/domain/events"
	svc "todoapp/services/notification-service/internal/service"
)

type Engine struct {
	created   *template.Template
	completed *template.Template
	deleted   *template.Template
	timeFmt   string
}

func NewEngine() (*Engine, error) {
	created, err := template.New("task_created").Parse(taskCreatedTemplate)
	if err != nil {
		return nil, err
	}
	completed, err := template.New("task_completed").Parse(taskCompletedTemplate)
	if err != nil {
		return nil, err
	}
	deleted, err := template.New("task_deleted").Parse(taskDeletedTemplate)
	if err != nil {
		return nil, err
	}

	return &Engine{
		created:   created,
		completed: completed,
		deleted:   deleted,
		timeFmt:   "02 Jan 2006 15:04",
	}, nil
}

func (e *Engine) RenderTaskCreated(event events.TaskEvent) (svc.TemplateResult, error) {
	return e.render(e.created, "Создана новая задача", event)
}

func (e *Engine) RenderTaskCompleted(event events.TaskEvent) (svc.TemplateResult, error) {
	return e.render(e.completed, "Задача выполнена", event)
}

func (e *Engine) RenderTaskDeleted(event events.TaskEvent) (svc.TemplateResult, error) {
	return e.render(e.deleted, "Задача удалена", event)
}

func (e *Engine) render(tpl *template.Template, subject string, event events.TaskEvent) (svc.TemplateResult, error) {
	var buf bytes.Buffer
	data := map[string]any{
		"Title":       event.Title,
		"Description": event.Description,
		"DueDate":     formatTime(event.DueDate, e.timeFmt),
	}
	if err := tpl.Execute(&buf, data); err != nil {
		return svc.TemplateResult{}, err
	}
	return svc.TemplateResult{Subject: subject, Body: buf.String()}, nil
}

func formatTime(value *time.Time, layout string) string {
	if value == nil {
		return "—"
	}
	return value.In(time.Local).Format(layout)
}

const taskCreatedTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><style>body{font-family:Arial,sans-serif;}h1{color:#2c3e50;}</style></head>
<body>
  <h1>Создана новая задача</h1>
  <p><strong>Название:</strong> {{.Title}}</p>
  <p><strong>Описание:</strong> {{.Description}}</p>
  <p><strong>Дедлайн:</strong> {{.DueDate}}</p>
</body>
</html>`

const taskCompletedTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><style>body{font-family:Arial,sans-serif;}h1{color:#27ae60;}</style></head>
<body>
  <h1>Задача выполнена!</h1>
  <p>Отличная работа! Задача <strong>{{.Title}}</strong> отмечена как завершенная.</p>
</body>
</html>`

const taskDeletedTemplate = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><style>body{font-family:Arial,sans-serif;}h1{color:#c0392b;}</style></head>
<body>
  <h1>Задача удалена</h1>
  <p>Задача <strong>{{.Title}}</strong> была удалена.</p>
  <p>{{.Description}}</p>
</body>
</html>`
