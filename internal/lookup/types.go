package lookup

type Category struct {
	ID   int16  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Location struct {
	ID      int16   `json:"id"`
	Name    string  `json:"name"`
	Address *string `json:"address"`
}
