package routes

type FormData struct {
	Values   map[string]string
	Errors   map[string][]string
	Messages []string
}

type RouteData struct {
	Template    string
	HttpStatus  int
	Title       string
	Description string
	FormData    FormData
}
