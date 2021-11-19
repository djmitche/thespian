package gen

type MailboxGenerator interface {
	GenerateGo(*formatter)
	ActorBuilderStructDecl() string
	ActorRxStructDecl() string
	ActorRxInitializer() string
	ActorTxStructDecl() string
	ActorTxInitializer() string
	ActorTxStructMethod() string
	ActorSpawnSetupClause() string
	ActorLoopCase() string
	ActorCleanupClause() string
}

func NewMailboxGeneratorForActor(thisPackage, actorName, fieldName, mboxTypeQual string, yml ActorMailboxYml) MailboxGenerator {
	switch yml.Kind {
	case "simple":
		return &SimpleMailboxGenerator{
			ThisPackage:  thisPackage,
			MessageType:  yml.MessageType,
			MboxTypeBase: yml.Type,
			MboxTypeQual: mboxTypeQual,
			ActorName:    actorName,
			FieldName:    fieldName,
		}
	case "ticker":
		return &TickerMailboxGenerator{
			ThisPackage:  thisPackage,
			MboxTypeBase: yml.Type,
			MboxTypeQual: mboxTypeQual,
			ActorName:    actorName,
			FieldName:    fieldName,
		}
	default:
		bail("unknown mailbox kind %s", yml.Kind)
		return nil
	}
}

func NewMailboxGeneratorForMailbox(thisPackage, typeName string, yml MailboxYml) MailboxGenerator {
	switch yml.Kind {
	case "simple":
		return &SimpleMailboxGenerator{
			ThisPackage:  thisPackage,
			MessageType:  yml.MessageType,
			MboxTypeBase: typeName,
		}
	case "ticker":
		return &TickerMailboxGenerator{
			ThisPackage:  thisPackage,
			MboxTypeBase: typeName,
		}
	default:
		bail("unknown template kind %s", yml.Kind)
		return nil
	}
}
