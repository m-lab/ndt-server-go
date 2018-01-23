package protocol

// Message types:

const KvCommFailure byte = 0
const KvSrvQueue byte = 1
const KvMsgLogin byte = 2
const KvTestPrepare byte = 3
const KvTestStart byte = 4
const KvTestMsg byte = 5
const KvTestFinalize byte = 6
const KvMsgError byte = 7
const KvMsgResults byte = 8
const KvMsgLogout byte = 9
const KvMsgWaiting byte = 10
const KvMsgExtendedLogin byte = 11

// Test identifiers:

const KvTestMid int = 1
const KvTestC2s int = 2
const KvTestS2c int = 4
const KvTestSfw int = 8
const KvTestStatus int = 16
const KvTestMeta int = 32
const KvTestC2sExt int = 64
const KvTestS2cExt int = 128

// Errors returned to client:

const KvSrvQueueHeartbeat string = "9990"
const KvSrvQueueServerFault string = "9977"
const KvSrvQueueServerBusy string = "9987"
const KvSrvQueueServerBusy60s string = "9999"
