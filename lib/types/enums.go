package types

type AccessType int

const (
	AccessTypeUnknown AccessType = iota
	AccessTypeUrl
	AccessTypeFile
	AccessTypeFileExtension
	AccessTypeContentType
	AccessTypeDomain
)

func (at AccessType) String() string {
	switch at {
	case AccessTypeUrl:
		return "url"
	case AccessTypeFile:
		return "file"
	case AccessTypeFileExtension:
		return "file extension"
	case AccessTypeContentType:
		return "content type"
	case AccessTypeDomain:
		return "domain"
	default:
		return "unknown"
	}
}
