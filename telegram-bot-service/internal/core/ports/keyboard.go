package ports

type Button struct {
	Text         string
	CallbackData string
	URL          string
	WebAppURL    string
}

type Keyboard struct {
	Rows [][]Button
}

func NewKeyboard(rows ...[]Button) *Keyboard {
	return &Keyboard{Rows: rows}
}

func ButtonRow(buttons ...Button) []Button {
	return buttons
}

func CallbackButton(text, data string) Button {
	return Button{Text: text, CallbackData: data}
}

func URLButton(text, url string) Button {
	return Button{Text: text, URL: url}
}

func WebAppButton(text, url string) Button {
	return Button{Text: text, WebAppURL: url}
}
