package gen

type MailboxGenerator interface {
	GenerateGo(*formatter)
	ActorPublicStructDecl() string
	ActorPublicStructMethod() string
	ActorSpawnSetupClause() string
	ActorSpawnRxAssignmentClause() string
	ActorSpawnHandleInitializer() string
	ActorLoopCase() string
	ActorCleanupClause() string
}

func NewMailboxGeneratorForActor(thisPackage, actorName, fieldName string, yml ActorMailboxYml) MailboxGenerator {
	switch yml.Kind {
	case "simple":
		return &SimpleMailboxGenerator{
			ThisPackage: thisPackage,
			MessageType: yml.MessageType,
			Package:     yml.Package,
			Type:        yml.Type,
			ActorName:   actorName,
			FieldName:   fieldName,
		}
	case "ticker":
		return &TickerMailboxGenerator{
			ThisPackage: thisPackage,
			Package:     yml.Package,
			Type:        yml.Type,
			ActorName:   actorName,
			FieldName:   fieldName,
		}
	default:
		bail("unknown template kind %s", yml.Kind)
		return nil
	}
}

func NewMailboxGeneratorForMailbox(thisPackage, typeName string, yml MailboxYml) MailboxGenerator {
	switch yml.Kind {
	case "simple":
		return &SimpleMailboxGenerator{
			ThisPackage: thisPackage,
			MessageType: yml.MessageType,
			Type:        typeName,
		}
	case "ticker":
		return &TickerMailboxGenerator{
			ThisPackage: thisPackage,
			Type:        typeName,
		}
	default:
		bail("unknown template kind %s", yml.Kind)
		return nil
	}
}
