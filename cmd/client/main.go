package main

import (
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type statusMsg int

type errMsg struct{ error }

type App struct {
	err    error
	status int
}

func (app App) Init() tea.Cmd {
	return checkServer
}

func (app App) View() string {
	if app.status != 0 {
		return fmt.Sprintf("Done! Got: %d. Press CTRL+C to exit", app.status)
	}
	if app.err != nil {
		return fmt.Sprintf("something went wrong: %s", app.err)
	}
	return "Processing. Press CTRL+C to exit"
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return app, tea.Quit
		}
	case errMsg:
		app.err = msg
	case statusMsg:
		app.status = int(msg)
	}
	return app, nil
}

func main() {
	app := App{}
	prog := tea.NewProgram(app)

	prog.Run()
}

var url = "http://google.com"

func checkServer() tea.Msg {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := c.Get(url)
	if err != nil {
		return errMsg{err}
	}
	defer res.Body.Close() // nolint:errcheck

	return statusMsg(res.StatusCode)
}
