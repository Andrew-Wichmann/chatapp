package main

import (
	"fmt"
	"strings"

	"github.com/Andrew-Wichmann/chatapp/pkg/client"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ error }

type App struct {
	err          error
	messageInput textarea.Model
	chatHistory  viewport.Model
    userNameDialog textarea.Model
	rpcClient    client.ChatAppClient
    loggedIn     bool
	history      []string
    userName     string
}

func newApp() App {
	ta := textarea.New()

	ta.Placeholder = "Start chatting"
	ta.Prompt = "┃ "
	ta.ShowLineNumbers = false
	ta.SetHeight(1)
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 5)

    // NOTE: there might be a better way to disable these keybinds
	vp.SetContent("Welcome to the chatroom")
    vp.KeyMap.PageDown.SetEnabled(false)
    vp.KeyMap.PageUp.SetEnabled(false)
    vp.KeyMap.HalfPageUp.SetEnabled(false)
    vp.KeyMap.HalfPageDown.SetEnabled(false)
    vp.KeyMap.Up.SetEnabled(false)
    vp.KeyMap.Down.SetEnabled(false)

    und := textarea.New()
    und.Placeholder = "<enter username here>"
	und.Prompt = "┃ "
	und.ShowLineNumbers = false
	und.SetHeight(1)
	und.Focus()
	und.KeyMap.InsertNewline.SetEnabled(false)

	return App{
		messageInput: ta,
		chatHistory:  vp,
        userNameDialog: und,
	}
}

func (app App) Init() tea.Cmd {
	return textarea.Blink
}

func (app App) View() string {
	if app.err != nil {
		return fmt.Sprintf("something went wrong: %s", app.err)
	}
    if app.loggedIn == false {
        return fmt.Sprintf("What should we call you?\n%s", app.userNameDialog.View())
    }
	return fmt.Sprintf("%s\n%s\nPress CTRL+C to exit", app.chatHistory.View(), app.messageInput.View())
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if app.loggedIn == false {
        if v, ok := msg.(tea.KeyMsg); ok{
            if v.Type == tea.KeyEnter {
                app.userName = app.userNameDialog.Value()
                app.loggedIn = true
                app.messageInput.Focus()
            }
            if v.Type == tea.KeyCtrlC {
                return app, tea.Quit
            }
        }
        var userNameDialogCmd tea.Cmd
        app.userNameDialog, userNameDialogCmd = app.userNameDialog.Update(msg)
        return app, userNameDialogCmd
    }

	var (
		messageCmd        tea.Cmd
		chatHistoryCmd    tea.Cmd
	)

	app.messageInput, messageCmd = app.messageInput.Update(msg)
	app.chatHistory, chatHistoryCmd = app.chatHistory.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return app, tea.Quit
		case tea.KeyEnter:
			cmd := sendMessage(app)
			app.messageInput.Reset()
			return app, cmd
		}
	case errMsg:
		app.err = msg
	case client.ChatResponse:
        app.history = append(app.history, fmt.Sprintf("%s: %s", msg.Username, msg.Message))
		app.chatHistory.SetContent(strings.Join(app.history, "\n"))
        app.chatHistory.GotoBottom()
	}
	return app, tea.Batch(messageCmd, chatHistoryCmd)
}

func main() {
	prog := tea.NewProgram(newApp())

	prog.Run()
}


func sendMessage(app App) tea.Cmd {
	return func() tea.Msg {
        msg := client.ChatMessage{Message: app.messageInput.Value(), Username: app.userName}
		response, err := app.rpcClient.SendMessageRPC(msg)
		if err != nil {
			return errMsg{err}
		}
		return response
	}
}
