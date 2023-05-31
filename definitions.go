/*
package mqtt implements MQTT v3.1.1 protocol providing users of this package with
low level decoding and encoding primitives and complete documentation sufficient
to grapple with the concepts of the MQTT protocol.

If you are new to MQTT start by reading definitions.go.
*/
package mqtt

const (
	// Accepted protocol level as per MQTT v3.1.1. This goes in the CONNECT variable header.
	DefaultProtocolLevel = 4
	// Accepted protocol as per MQTT v3.1.1. This goes in the CONNECT variable header.
	DefaultProtocol = "MQTT"
	// Size on wire after being encoded.
	maxRemainingLengthSize = 4
	// Max value Remaining Length can take 0xfff_ffff. When encoded over the wire this value yields 0xffff_ff7f.
	maxRemainingLengthValue = 0xfff_ffff
)

// Reserved flags for PUBREL, SUBSCRIBE and UNSUBSCRIBE packet types.
// This is effectively a PUBLISH flag with QoS1 set and no DUP or RETAIN bits.
const PacketFlagsPubrelSubUnsub PacketFlags = 0b10

// PacketType represents the 4 MSB bits in the first byte in an MQTT fixed header.
// takes on values 1..14. PacketType and PacketFlags are present in all MQTT packets.
type PacketType byte

const (
	// 0 Forbidden/Reserved
	_ PacketType = iota
	// A CONNECT packet is sent from Client to Server, it is a Client request to connect to a Server.
	// After a network connection is established by a client to a server at the transport layer, the first
	// packet sent from the client to the server must be a Connect packet.
	// A Client can only send the CONNECT Packet once over a Network Connection.
	// The CONNECT packet contains a 10 byte variable header and a
	// payload determined by flags present in variable header. See [VariablesConnect]. 0x10.
	PacketConnect
	// The CONNACK Packet is the packet sent by the Server in response to a CONNECT Packet received from a Client.
	// The first packet sent from the Server to the Client MUST be a CONNACK Packet
	// The payload contains a 2 byte variable header and no payload. 0x20.
	PacketConnack
	// A PUBLISH Control Packet is sent from a Client to a Server or from Server to a Client to transport an Application Message.
	// It's payload contains a variable header with a MQTT encoded string for the topic name and a packet identifier.
	// The payload may or may not contain a Application Message that is being published. The length of this Message
	// can be calculated by subtracting the length of the variable header from the Remaining Length field that is in the Fixed Header. 0x3?.
	PacketPublish
	// A PUBACK Packet is the response to a PUBLISH Packet with QoS level 1. It's Variable header contains the packet identifier. No payload. 0x40.
	PacketPuback
	// A PUBREC Packet is the response to a PUBLISH Packet with QoS 2. It is the second packet of the QoS 2 protocol exchange. It's Variable header contains the packet identifier. No payload. 0x50.
	PacketPubrec
	// A PUBREL Packet is the response to a PUBREC Packet. It is the third packet of the QoS 2 protocol exchange. It's Variable header contains the packet identifier. No payload. 0x62.
	PacketPubrel
	// The PUBCOMP Packet is the response to a PUBREL Packet. It is the fourth and final packet of the QoS 2 protocol exchange. It's Variable header contains the packet identifier. No payload. 0x70.
	PacketPubcomp
	// The SUBSCRIBE Packet is sent from the Client to the Server to create one or more Subscriptions.
	// Each Subscription registers a Client’s interest in one or more Topics. The Server sends PUBLISH
	// Packets to the Client in order to forward Application Messages that were published to Topics that match these Subscriptions.
	// The SUBSCRIBE Packet also specifies (for each Subscription) the maximum QoS with which the Server can
	// send Application Messages to the Client.
	// The variable header of a subscribe topic contains the packet identifier. The payload contains a list of topic filters, see [VariablesSubscribe]. 0x82.
	PacketSubscribe
	// A SUBACK Packet is sent by the Server to the Client to confirm receipt and processing of a SUBSCRIBE Packet.
	// The variable header contains the packet identifier. The payload contains a list of octet return codes for each subscription requested by client, see [VariablesSuback]. 0x90.
	PacketSuback
	// An UNSUBSCRIBE Packet is sent by the Client to the Server, to unsubscribe from topics.
	// The variable header contains the packet identifier. Its payload contains a list of mqtt encoded strings corresponding to unsubscribed topics, see [VariablesUnsubscribe]. 0xa2.
	PacketUnsubscribe
	// The UNSUBACK Packet is sent by the Server to the Client to confirm receipt of an UNSUBSCRIBE Packet.
	// The variable header contains the packet identifier. It has no payload. 0xb0.
	PacketUnsuback
	// The PINGREQ Packet is sent from a Client to the Server. It can be used to:
	//  - Indicate to the Server that the Client is alive in the absence of any other Control Packets being sent from the Client to the Server.
	//  - Request that the Server responds to confirm that it is alive.
	//  - Exercise the network to indicate that the Network Connection is active.
	// No payload or variable header. 0xc0.
	PacketPingreq
	// A PINGRESP Packet is sent by the Server to the Client in response to a PINGREQ Packet. It indicates that the Server is alive.
	// No payload or variable header. 0xd0.
	PacketPingresp
	// The DISCONNECT Packet is the final Control Packet sent from the Client to the Server. It indicates that the Client is disconnecting cleanly.
	// No payload or variable header. 0xe0.
	PacketDisconnect
)

// QoSLevel represents the Quality of Service specified by the client.
// The server can choose to provide or reject requested QoS. The values
// of QoS range from 0 to 2, each representing a different methodology for
// message delivery guarantees.
type QoSLevel uint8

// QoS indicates the level of assurance for packet delivery.
const (
	// QoS0 at most once delivery. Arrives either once or not at all. Depends on capabilities of underlying network.
	QoS0 QoSLevel = iota
	// QoS1 at least once delivery. Ensures message arrives at receiver at least once.
	QoS1
	// QoS2 Exactly once delivery. Highest quality service. For use when neither loss nor duplication of messages are acceptable.
	// There is an increased overhead associated with this quality of service.
	QoS2
	// Reserved, must not be used.
	reservedQoS3
	// QoSSubfail marks a failure in SUBACK. This value cannot be encoded into a header
	// and is only returned upon an unsuccessful subscribe to a topic in an SUBACK packet.
	QoSSubfail QoSLevel = 0x80
)

// ConnectReturnCode represents the CONNACK return code, which is the second byte in the variable header.
// It indicates if the connection was successful (0 value) or if the connection attempt failed on the server side.
// ConnectReturnCode also implements the error interface and can be returned on a failed connection.
type ConnectReturnCode uint8

const (
	ReturnCodeConnAccepted ConnectReturnCode = iota
	ReturnCodeUnnaceptableProtocol
	ReturnCodeIdentifierRejected
	ReturnCodeServerUnavailable
	ReturnCodeBadUserCredentials
	ReturnCodeUnauthorized
	minInvalidReturnCode
)

// Error implements the error interface for a non-zero return code.
func (rc ConnectReturnCode) Error() string { return rc.String() }
