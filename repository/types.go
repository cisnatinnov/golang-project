package repository

type Estate struct {
	Id     string
	Length int
	Width  int
}

type Tree struct {
	Id       string
	EstateId string
	X        int
	Y        int
	Height   int
}

type CreateEstateInput struct {
	Length int
	Width  int
}

type CreateEstateOutput struct {
	Id string
}

type CreateTreeInput struct {
	EstateId string
	X        int
	Y        int
	Height   int
}

type CreateTreeOutput struct {
	Id string
}

type GetEstateStatsInput struct {
	EstateId string
}

type GetEstateStatsOutput struct {
	Count  int
	Max    int
	Min    int
	Median int
}

type GetTreesByEstateIdInput struct {
	EstateId string
}

type GetTreesByEstateIdOutput struct {
	Trees []Tree
}
