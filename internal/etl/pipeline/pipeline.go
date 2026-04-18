package pipeline

type Definition struct {
	ID            string
	Name          string
	SourceType    string
	TargetType    string
	TransformList []string
}
