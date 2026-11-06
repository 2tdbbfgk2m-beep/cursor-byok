// inbound.go 实现上行协议的解码、摘要与命令类型识别。
package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"cursor/gen/agentv1"
	"cursor/gen/aiserverv1"
	"cursor/internal/backend/agent/core"
	"cursor/internal/logger"

	"google.golang.org/protobuf/proto"
)

// ReadAppendRequestID 从 BidiAppendRequest 中读取 request_id 文本。
func ReadAppendRequestID(input *aiserverv1.BidiAppendRequest) string {
	if input == nil {
		return ""
	}
	return ReadBidiRequestID(input.GetRequestId())
}

// ReadBidiRequestID 从 BidiRequestId 结构中提取并去除首尾空白。
func ReadBidiRequestID(input *aiserverv1.BidiRequestId) string {
	if input == nil {
		return ""
	}
	return strings.TrimSpace(input.GetRequestId())
}

// NormalizeRequestID 规范化请求标识并去除首尾空白。
func NormalizeRequestID(requestID string) string {
	return strings.TrimSpace(requestID)
}

// DecodeAgentClientMessage 解析 hex 文本为 AgentClientMessage，并返回消息类型标签。
func DecodeAgentClientMessage(hexData string) (*agentv1.AgentClientMessage, string, error) {
	trimmed := strings.TrimSpace(hexData)
	if trimmed == "" {
		return nil, "", nil
	}
	payload, err := hex.DecodeString(trimmed)
	if err != nil {
		return nil, "", fmt.Errorf("bidi append data is not valid hex: %w", err)
	}
	clientMessage := &agentv1.AgentClientMessage{}
	if err := proto.Unmarshal(payload, clientMessage); err != nil {
		return nil, "", fmt.Errorf("decode agent client message failed: %w", err)
	}
	clientKind := detectClientMessageKind(clientMessage)
	if clientKind == "" {
		logUnknownClientMessage(payload, clientMessage)
	}
	return clientMessage, clientKind, nil
}

// MapClientMessageToCommandKind 将上行协议消息映射为运行时命令类型。
func MapClientMessageToCommandKind(message *agentv1.AgentClientMessage, clientKind string) (runtimecore.CommandKind, error) {
	switch strings.TrimSpace(clientKind) {
	case "run_request":
		return runtimecore.CommandKindRunRequested, nil
	case "prewarm_request":
		return runtimecore.CommandKindPrewarmRequested, nil
	case "conversation_action":
		if message == nil || message.GetConversationAction() == nil {
			return "", fmt.Errorf("conversation_action payload is required")
		}
		switch message.GetConversationAction().GetAction().(type) {
		case *agentv1.ConversationAction_CancelAction:
			return runtimecore.CommandKindCancelRequested, nil
		case *agentv1.ConversationAction_UserMessageAction,
			*agentv1.ConversationAction_ResumeAction,
			*agentv1.ConversationAction_SummarizeAction,
			*agentv1.ConversationAction_ShellCommandAction,
			*agentv1.ConversationAction_StartPlanAction,
			*agentv1.ConversationAction_ExecutePlanAction,
			*agentv1.ConversationAction_AsyncAskQuestionCompletionAction,
			*agentv1.ConversationAction_CancelSubagentAction,
			*agentv1.ConversationAction_BackgroundShellAction,
			*agentv1.ConversationAction_BackgroundTaskCompletionAction:
			return runtimecore.CommandKindConversationActionRecordOnly, nil
		default:
			return "", fmt.Errorf("unsupported conversation_action payload")
		}
	case "exec_client_message":
		return runtimecore.CommandKindExecClientMessage, nil
	case "interaction_response":
		return runtimecore.CommandKindInteractionResponse, nil
	case "exec_client_control_message":
		return runtimecore.CommandKindExecClientControlMessage, nil
	case "client_heartbeat":
		return runtimecore.CommandKindClientHeartbeat, nil
	case "kv_client_message":
		return runtimecore.CommandKindKVClientMessage, nil
	default:
		return "", fmt.Errorf("unsupported client message kind: %s", clientKind)
	}
}

// IsResumeRunRequest 判断当前消息是否为带 `resume_action` 的 `run_request`。
func IsResumeRunRequest(message *agentv1.AgentClientMessage) bool {
	if message == nil || message.GetRunRequest() == nil {
		return false
	}
	action := message.GetRunRequest().GetAction()
	if action == nil {
		return false
	}
	_, ok := action.GetAction().(*agentv1.ConversationAction_ResumeAction)
	return ok
}

// BuildClientHistoryEntry 将消息类型与负载摘要拼接成会话历史记录文本。
func BuildClientHistoryEntry(kind string, message *agentv1.AgentClientMessage) string {
	normalizedKind := strings.TrimSpace(kind)
	if normalizedKind == "" {
		normalizedKind = "unknown"
	}

	payload := extractClientMessagePayload(message)
	summary := summarizePayload(payload)
	if summary == "" {
		if normalizedKind == "unknown" {
			return ""
		}
		return normalizedKind
	}
	return fmt.Sprintf("%s:%s", normalizedKind, summary)
}

// detectClientMessageKind 判断 oneof message 当前承载的消息分支类型。
func detectClientMessageKind(message *agentv1.AgentClientMessage) string {
	if message == nil || message.GetMessage() == nil {
		return ""
	}
	switch message.GetMessage().(type) {
	case *agentv1.AgentClientMessage_RunRequest:
		return "run_request"
	case *agentv1.AgentClientMessage_PrewarmRequest:
		return "prewarm_request"
	case *agentv1.AgentClientMessage_ConversationAction:
		return "conversation_action"
	case *agentv1.AgentClientMessage_ExecClientMessage:
		return "exec_client_message"
	case *agentv1.AgentClientMessage_InteractionResponse:
		return "interaction_response"
	case *agentv1.AgentClientMessage_ExecClientControlMessage:
		return "exec_client_control_message"
	case *agentv1.AgentClientMessage_ClientHeartbeat:
		return "client_heartbeat"
	case *agentv1.AgentClientMessage_KvClientMessage:
		return "kv_client_message"
	default:
		return ""
	}
}

// extractClientMessagePayload 从 oneof 分支中提取原始 bytes 负载。
func extractClientMessagePayload(message *agentv1.AgentClientMessage) []byte {
	if message == nil || message.GetMessage() == nil {
		return nil
	}

	switch item := message.GetMessage().(type) {
	case *agentv1.AgentClientMessage_RunRequest:
		return marshalProtoMessage(item.RunRequest)
	case *agentv1.AgentClientMessage_PrewarmRequest:
		return marshalProtoMessage(item.PrewarmRequest)
	case *agentv1.AgentClientMessage_ConversationAction:
		return marshalProtoMessage(item.ConversationAction)
	case *agentv1.AgentClientMessage_ExecClientMessage:
		return marshalProtoMessage(item.ExecClientMessage)
	case *agentv1.AgentClientMessage_InteractionResponse:
		return marshalProtoMessage(item.InteractionResponse)
	case *agentv1.AgentClientMessage_ExecClientControlMessage:
		return marshalProtoMessage(item.ExecClientControlMessage)
	case *agentv1.AgentClientMessage_ClientHeartbeat:
		return marshalProtoMessage(item.ClientHeartbeat)
	case *agentv1.AgentClientMessage_KvClientMessage:
		return marshalProtoMessage(item.KvClientMessage)
	default:
		return nil
	}
}

// marshalProtoMessage 将 proto 消息重新编码为 bytes，用于调试摘要展示。
func marshalProtoMessage(message proto.Message) []byte {
	if message == nil {
		return nil
	}
	payload, err := proto.Marshal(message)
	if err != nil {
		return nil
	}
	return payload
}

// summarizePayload 生成可读摘要：优先文本，无法直接读时回退为 hex 片段。
func summarizePayload(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}

	if utf8.Valid(payload) {
		text := strings.TrimSpace(string(payload))
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.ReplaceAll(text, "\r", " ")
		text = strings.TrimSpace(text)
		if text != "" {
			return truncateText(text, 120)
		}
	}
	return "hex:" + truncateText(hex.EncodeToString(payload), 120)
}

// truncateText 按 rune 数截断文本，避免在多字节字符中间截断导致乱码。
func truncateText(text string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes]) + "..."
}

// unknownKindLogOnce 记录已打印过的 unknown 消息签名，避免重复日志刷屏。
var unknownKindLogOnce sync.Map

// logUnknownClientMessage 在上行 AgentClientMessage 解不出已知 oneof 种类时，
// 打印 unknown fields 的字段号与 wire type，用于反查新版客户端新增的 oneof 成员。
func logUnknownClientMessage(payload []byte, message *agentv1.AgentClientMessage) {
	oneofType := fmt.Sprintf("%T", message.GetMessage())
	unknown := message.ProtoReflect().GetUnknown()
	var fieldDescs []string
	if len(unknown) > 0 {
		fieldDescs = describeUnknownFields(unknown)
	} else {
		fieldDescs = []string{"none"}
	}
	signature := oneofType + "|" + strings.Join(fieldDescs, ",")
	if _, loaded := unknownKindLogOnce.LoadOrStore(signature, struct{}{}); loaded {
		return
	}
	logger.Infof(
		"unknown agent client message kind: oneof=%s unknown_fields=[%s] payload_len=%d payload_head=%s",
		oneofType,
		strings.Join(fieldDescs, ","),
		len(payload),
		truncateText(hex.EncodeToString(payload), 160),
	)
}

// describeUnknownFields 解析 proto unknown fields 原始字节，返回 "field=<n> wire=<t> len=<l>" 描述列表。
func describeUnknownFields(unknown []byte) []string {
	var descs []string
	rest := unknown
	for len(rest) > 0 {
		tag, n := binary.Uvarint(rest)
		if n <= 0 {
			descs = append(descs, "malformed-tag")
			break
		}
		rest = rest[n:]
		fieldNum := tag >> 3
		wireType := tag & 0x7
		length := 0
		switch wireType {
		case 0: // varint
			_, n = binary.Uvarint(rest)
			if n > 0 {
				length = n
				rest = rest[n:]
			}
		case 1: // 64-bit
			length = 8
			if len(rest) >= 8 {
				rest = rest[8:]
			}
		case 2: // length-delimited
			l, n := binary.Uvarint(rest)
			if n > 0 {
				length = int(l)
				if int(n)+int(l) <= len(rest) {
					rest = rest[n+int(l):]
				} else {
					rest = nil
				}
			}
		case 5: // 32-bit
			length = 4
			if len(rest) >= 4 {
				rest = rest[4:]
			}
		default:
			descs = append(descs, fmt.Sprintf("field=%d wire=%d", fieldNum, wireType))
			continue
		}
		descs = append(descs, fmt.Sprintf("field=%d wire=%d len=%d", fieldNum, wireType, length))
	}
	return descs
}
