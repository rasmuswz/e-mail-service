package commandprotocol

type Command int;
const (
	CMD_MTA_PROVIDER_SHUTDOWN = 0x00;
	CMD_MTA_PROVIDER_NOTIFY_DOWN = 0x01;
	CMD_MTA_PROVIDER_NOTIFY_UP = 0x02;
)