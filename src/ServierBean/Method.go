package ServierBean

type Method interface {
	ParseDate(data []byte)

	ReturnMessage() (rsBytes []byte)
}
