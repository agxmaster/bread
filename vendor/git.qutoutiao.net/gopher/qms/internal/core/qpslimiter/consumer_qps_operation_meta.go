package qpslimiter

// ConsumerKeys contain consumer keys
type ConsumerKeys struct {
	MicroServiceName string
	OperationID      string
	//SchemaQualifiedName    string
	OperationQualifiedName string // Deprecated
}

// ProviderKeys contain provider keys
type ProviderKeys struct {
	MicroServiceName string
	OperationID      string
}

//Prefix is const
const Prefix = "qms.flowcontrol"

// GetConsumerKey get specific key for consumer
func GetConsumerKey(serviceName, operationID string) *ConsumerKeys {
	keys := new(ConsumerKeys)
	//for mesher to govern
	if serviceName != "" {
		keys.MicroServiceName = serviceName
	}
	//if schemaID != "" {
	//	keys.SchemaQualifiedName = strings.Join([]string{keys.MicroServiceName, schemaID}, ".")
	//}
	if operationID != "" {
		keys.OperationID = operationID
	}
	return keys
}

// GetProviderKey get specific key for provider
func GetProviderKey(serviceName, operationID string) *ProviderKeys {
	keys := &ProviderKeys{}
	if serviceName != "" {
		keys.MicroServiceName = serviceName
	}

	if operationID != "" {
		keys.OperationID = operationID
	}

	return keys
}

// GetMicroServiceSchemaOpQualifiedName get micro-service schema operation qualified name
func (op *ConsumerKeys) GetMicroServiceSchemaOpQualifiedName() string {
	return op.OperationQualifiedName
}

// GetMicroServiceName get micro-service name
func (op *ConsumerKeys) GetMicroServiceName() string {
	return op.MicroServiceName
}
