package http2

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/summerwind/protospec/log"
	"github.com/summerwind/protospec/protocol/action"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const (
	ProtocolType = "http2"

	DefaultMaxFrameSize        uint32 = 16384
	DefaultMaxFieldValueLength uint32 = 4096

	ActionSendData              = "http2.send_data"
	ActionSendDataFrame         = "http2.send_data_frame"
	ActionSendHeadersFrame      = "http2.send_headers_frame"
	ActionSendPriorityFrame     = "http2.send_priority_frame"
	ActionSendRSTStreamFrame    = "http2.send_rst_stream_frame"
	ActionSendSettingsFrame     = "http2.send_settings_frame"
	ActionSendPingFrame         = "http2.send_ping_frame"
	ActionSendGoAwayFrame       = "http2.send_goaway_frame"
	ActionSendWindowUpdateFrame = "http2.send_window_update_frame"
	ActionSendContinuationFrame = "http2.send_continuation_frame"
	ActionWaitHeadersFrame      = "http2.wait_headers_frame"
	ActionWaitRSTStreamFrame    = "http2.wait_rst_stream_frame"
	ActionWaitSettingsFrame     = "http2.wait_settings_frame"
	ActionWaitPingFrame         = "http2.wait_ping_frame"
	ActionWaitGoAwayFrame       = "http2.wait_goaway_frame"
	ActionWaitConnectionError   = "http2.wait_connection_error"
	ActionWaitConnectionClose   = "http2.wait_connection_close"
	ActionWaitStreamError       = "http2.wait_stream_error"
	ActionWaitStreamClose       = "http2.wait_stream_close"
)

var settingID = map[string]http2.SettingID{
	"SETTINGS_HEADER_TABLE_SIZE":      http2.SettingHeaderTableSize,
	"SETTINGS_ENABLE_PUSH":            http2.SettingEnablePush,
	"SETTINGS_MAX_CONCURRENT_STREAMS": http2.SettingMaxConcurrentStreams,
	"SETTINGS_INITIAL_WINDOW_SIZE":    http2.SettingInitialWindowSize,
	"SETTINGS_MAX_FRAME_SIZE":         http2.SettingMaxFrameSize,
	"SETTINGS_MAX_HEADER_LIST_SIZE":   http2.SettingMaxHeaderListSize,
	"SETTINGS_UNKNOWN":                0x99,
}

var errorCode = map[string]http2.ErrCode{
	"NO_ERROR":            http2.ErrCodeNo,
	"PROTOCOL_ERROR":      http2.ErrCodeProtocol,
	"INTERNAL_ERROR":      http2.ErrCodeInternal,
	"FLOW_CONTROL_ERROR":  http2.ErrCodeFlowControl,
	"SETTINGS_TIMEOUT":    http2.ErrCodeSettingsTimeout,
	"STREAM_CLOSED":       http2.ErrCodeStreamClosed,
	"FRAME_SIZE_ERROR":    http2.ErrCodeFrameSize,
	"REFUSED_STREAM":      http2.ErrCodeRefusedStream,
	"CANCEL":              http2.ErrCodeCancel,
	"COMPRESSION_ERROR":   http2.ErrCodeCompression,
	"CONNECT_ERROR":       http2.ErrCodeConnect,
	"ENHANCE_YOUR_CALM":   http2.ErrCodeEnhanceYourCalm,
	"INADEQUATE_SECURITY": http2.ErrCodeInadequateSecurity,
	"HTTP_1_1_REQUIRED":   http2.ErrCodeHTTP11Required,
	"UNKNOWN_ERROR":       0x99,
}

func init() {
	rand.Seed(time.Now().Unix())
}

type Setting struct {
	ID    string `json:"id"`
	Value uint32 `json:"value"`
}

type Field struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive"`
}

type Priority struct {
	StreamDependency uint32 `json:"stream_dependency"`
	Exclusive        bool   `json:"exclusive"`
	Weight           uint8  `json:"weight"`
}

type InitParam struct {
	Handshake           bool      `json:"handshake"`
	Settings            []Setting `json:"settings"`
	MaxFieldValueLength uint32    `json:"max_field_value_length"`
}

func (p *InitParam) Validate() error {
	for _, setting := range p.Settings {
		_, ok := settingID[setting.ID]
		if !ok {
			return fmt.Errorf("invalid setting ID: %s", setting.ID)
		}
	}

	return nil
}

type SendDataParam struct {
	Data []string `json:"data"`
}

func (p *SendDataParam) Validate() error {
	if len(strings.Join(p.Data, "")) == 0 {
		return errors.New("'data' must be specified")
	}

	return nil
}

type SendDataFrameParam struct {
	StreamID         uint32 `json:"stream_id"`
	EndStream        bool   `json:"end_stream"`
	PadLength        uint8  `json:"pad_length"`
	Data             string `json:"data"`
	DataLength       uint32 `json:"data_length"`
	FillMaxFrameSize bool   `json:"fill_max_frame_size"`
}

func (p *SendDataFrameParam) Validate() error {
	if p.Data != "" && p.DataLength != 0 {
		return errors.New("'data' and 'data length' cannot be specified at the same time")
	}

	return nil
}

type SendHeadersFrameParam struct {
	StreamID         uint32    `json:"stream_id"`
	EndStream        bool      `json:"end_stream"`
	EndHeaders       bool      `json:"end_headers"`
	PadLength        uint8     `json:"pad_length"`
	HeaderFields     []Field   `json:"header_fields"`
	NoDefaultFields  bool      `json:"no_default_fields"`
	FillMaxFrameSize bool      `json:"fill_max_frame_size"`
	Priority         *Priority `json:"priority"`
}

func (p *SendHeadersFrameParam) Validate() error {
	if len(p.HeaderFields) == 0 && p.NoDefaultFields && !p.FillMaxFrameSize {
		return errors.New("'header_fields' must contain at least one field")
	}

	return nil
}

type SendPriorityFrameParam struct {
	Priority
	StreamID uint32 `json:"stream_id"`
}

func (p *SendPriorityFrameParam) Validate() error {
	return nil
}

type SendRSTStreamFrameParam struct {
	StreamID  uint32 `json:"stream_id"`
	ErrorCode string `json:"error_code"`
}

func (p *SendRSTStreamFrameParam) Validate() error {
	_, ok := errorCode[p.ErrorCode]
	if !ok {
		return fmt.Errorf("invalid error code: %s", p.ErrorCode)
	}

	return nil
}

type SendSettingsFrameParam struct {
	Ack      bool      `json:"ack"`
	Settings []Setting `json:"settings"`
}

func (p *SendSettingsFrameParam) Validate() error {
	for _, setting := range p.Settings {
		_, ok := settingID[setting.ID]
		if !ok {
			return fmt.Errorf("invalid setting ID: %s", setting.ID)
		}
	}

	return nil
}

type SendPingFrameParam struct {
	Ack  bool   `json:"ack"`
	Data string `json:"data"`
}

func (p *SendPingFrameParam) Validate() error {
	return nil
}

type SendGoAwayFrameParam struct {
	LastStreamID        uint32 `json:"last_stream_id"`
	ErrorCode           string `json:"error_code"`
	AdditionalDebugData string `json:"additional_debug_data"`
}

func (p *SendGoAwayFrameParam) Validate() error {
	_, ok := errorCode[p.ErrorCode]
	if !ok {
		return fmt.Errorf("invalid error code: %s", p.ErrorCode)
	}

	return nil
}

type SendWindowUpdateFrameParam struct {
	StreamID            uint32 `json:"stream_id"`
	WindowSizeIncrement uint32 `json:"window_size_increment"`
}

func (p *SendWindowUpdateFrameParam) Validate() error {
	return nil
}

type SendContinuationFrameParam struct {
	StreamID     uint32  `json:"stream_id"`
	EndHeaders   bool    `json:"end_headers"`
	HeaderFields []Field `json:"header_fields"`
}

func (p *SendContinuationFrameParam) Validate() error {
	if len(p.HeaderFields) == 0 {
		return errors.New("'header_fields' must contain at least one field")
	}

	return nil
}

type WaitHeadersFrameParam struct {
	StreamID uint32 `json:"stream_id"`
}

func (p *WaitHeadersFrameParam) Validate() error {
	return nil
}

type WaitRSTStreamFrameParam struct {
	StreamID  uint32   `json:"stream_id"`
	ErrorCode []string `json:"error_code"`
}

func (p *WaitRSTStreamFrameParam) Validate() error {
	for _, code := range p.ErrorCode {
		_, ok := errorCode[code]
		if !ok {
			return fmt.Errorf("invalid error code: %s", code)
		}
	}

	return nil
}

type WaitSettingsFrameParam struct {
	Ack      bool              `json:"ack"`
	Settings map[string]uint32 `json:"settings"`
}

func (p *WaitSettingsFrameParam) Validate() error {
	if p.Ack && len(p.Settings) > 0 {
		return errors.New("settings cannot be specified when ack is 'true'")
	}

	for key, _ := range p.Settings {
		_, ok := settingID[key]
		if !ok {
			return fmt.Errorf("invalid setting ID: %s", key)
		}
	}

	return nil
}

type WaitPingFrameParam struct {
	Ack  bool   `json:"ack"`
	Data string `json:"data"`
}

func (p *WaitPingFrameParam) Validate() error {
	return nil
}

type WaitGoAwayFrameParam struct {
	LastStreamID uint32   `json:"last_stream_id"`
	ErrorCode    []string `json:"error_code"`
	DebugData    string   `json:"debug_data"`
}

func (p *WaitGoAwayFrameParam) Validate() error {
	for _, code := range p.ErrorCode {
		_, ok := errorCode[code]
		if !ok {
			return fmt.Errorf("invalid error code: %s", code)
		}
	}

	return nil
}

type WaitConnectionErrorParam struct {
	ErrorCode []string `json:"error_code"`
}

func (p *WaitConnectionErrorParam) Validate() error {
	for _, code := range p.ErrorCode {
		_, ok := errorCode[code]
		if !ok {
			return fmt.Errorf("invalid error code: %s", code)
		}
	}

	return nil
}

type WaitStreamErrorParam struct {
	StreamID  uint32   `json:"stream_id"`
	ErrorCode []string `json:"error_code"`
}

func (p *WaitStreamErrorParam) Validate() error {
	for _, code := range p.ErrorCode {
		_, ok := errorCode[code]
		if !ok {
			return fmt.Errorf("invalid error code: %s", code)
		}
	}

	return nil
}

type WaitStreamCloseParam struct {
	StreamID uint32 `json:"stream_id"`
}

func (p *WaitStreamCloseParam) Validate() error {
	if p.StreamID == 0 {
		return errors.New("stream_id must be greater than 0")
	}

	return nil
}

type Conn struct {
	net.Conn

	framer      *http2.Framer
	debugFramer *http2.Framer

	encoderBuf *bytes.Buffer
	encoder    *hpack.Encoder
	decoder    *hpack.Decoder

	settings            map[http2.SettingID]uint32
	maxFieldValueLength uint32

	server  bool
	timeout time.Duration
}

func NewConn(transport net.Conn) (*Conn, error) {
	debugBuf := new(bytes.Buffer)
	debugConn := io.MultiWriter(transport, debugBuf)

	framer := http2.NewFramer(debugConn, transport)
	framer.AllowIllegalWrites = true
	framer.AllowIllegalReads = true

	debugFramer := http2.NewFramer(debugBuf, debugBuf)
	debugFramer.AllowIllegalWrites = true
	debugFramer.AllowIllegalReads = true

	encoderBuf := new(bytes.Buffer)
	encoder := hpack.NewEncoder(encoderBuf)
	decoder := hpack.NewDecoder(4096, func(f hpack.HeaderField) {})

	settings := map[http2.SettingID]uint32{}

	return &Conn{
		Conn:                transport,
		framer:              framer,
		debugFramer:         debugFramer,
		encoderBuf:          encoderBuf,
		encoder:             encoder,
		decoder:             decoder,
		settings:            settings,
		maxFieldValueLength: DefaultMaxFieldValueLength,
	}, nil
}

func (conn *Conn) Init(param []byte) error {
	var p InitParam

	if err := json.Unmarshal(param, &p); err != nil {
		return err
	}

	if err := p.Validate(); err != nil {
		return err
	}

	if p.MaxFieldValueLength != 0 {
		conn.maxFieldValueLength = p.MaxFieldValueLength
	}

	if p.Handshake {
		if err := conn.handshake(p.Settings); err != nil {
			return err
		}
	}

	return nil
}

func (conn *Conn) Run(action string, param []byte) (interface{}, error) {
	switch action {
	case ActionSendData:
		return conn.sendData(param)
	case ActionSendDataFrame:
		return conn.sendDataFrame(param)
	case ActionSendHeadersFrame:
		return conn.sendHeadersFrame(param)
	case ActionSendPriorityFrame:
		return conn.sendPriorityFrame(param)
	case ActionSendRSTStreamFrame:
		return conn.sendRSTStreamFrame(param)
	case ActionSendSettingsFrame:
		return conn.sendSettingsFrame(param)
	case ActionSendPingFrame:
		return conn.sendPingFrame(param)
	case ActionSendGoAwayFrame:
		return conn.sendGoAwayFrame(param)
	case ActionSendWindowUpdateFrame:
		return conn.sendWindowUpdateFrame(param)
	case ActionSendContinuationFrame:
		return conn.sendContinuationFrame(param)
	case ActionWaitHeadersFrame:
		return conn.waitHeadersFrame(param)
	case ActionWaitRSTStreamFrame:
		return conn.waitRSTStreamFrame(param)
	case ActionWaitSettingsFrame:
		return conn.waitSettingsFrame(param)
	case ActionWaitPingFrame:
		return conn.waitPingFrame(param)
	case ActionWaitGoAwayFrame:
		return conn.waitGoAwayFrame(param)
	case ActionWaitConnectionError:
		return conn.waitConnectionError(param)
	case ActionWaitConnectionClose:
		return conn.waitConnectionClose(param)
	case ActionWaitStreamError:
		return conn.waitStreamError(param)
	case ActionWaitStreamClose:
		return conn.waitStreamClose(param)
	default:
		return nil, fmt.Errorf("invalid action: %s", action)
	}
}

func (conn *Conn) SetTimeout(timeout time.Duration) {
	conn.timeout = timeout
}

func (conn *Conn) SetMode(server bool) {
	conn.server = server
}

func (conn *Conn) MaxFrameSize() uint32 {
	val, ok := conn.settings[http2.SettingMaxFrameSize]
	if !ok {
		return DefaultMaxFrameSize
	}

	return val
}

func (conn *Conn) readFrame() (http2.Frame, error) {
	if err := conn.SetReadDeadline(time.Now().Add(conn.timeout)); err != nil {
		return nil, err
	}

	f, err := conn.framer.ReadFrame()
	if f != nil {
		log.Debug(fmt.Sprintf("recv: %s", getEvent(f)))
	}

	return f, err
}

func (conn *Conn) handshake(settings []Setting) error {
	var (
		initSettings []http2.Setting
		state        int
	)

	for _, setting := range settings {
		initSettings = append(initSettings, http2.Setting{
			ID:  settingID[setting.ID],
			Val: setting.Value,
		})
	}

	fmt.Fprintf(conn, http2.ClientPreface)
	logWrite("Connection Preface")

	if err := conn.framer.WriteSettings(initSettings...); err != nil {
		return err
	}
	conn.logWriteFrame()

	for {
		f, err := conn.readFrame()
		if err != nil {
			return err
		}

		sf, ok := f.(*http2.SettingsFrame)
		if !ok {
			continue
		}

		if !sf.IsAck() {
			sf.ForeachSetting(func(setting http2.Setting) error {
				conn.settings[setting.ID] = setting.Val
				return nil
			})

			if err := conn.framer.WriteSettingsAck(); err != nil {
				return err
			}
			conn.logWriteFrame()
		}

		state += 1
		if state > 1 {
			break
		}
	}

	return nil
}

func (conn *Conn) sendData(param []byte) (interface{}, error) {
	var p SendDataParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for _, data := range p.Data {
		var (
			buf []byte
			err error
		)

		if strings.HasPrefix(data, "0x") {
			buf, err = hex.DecodeString(data[2:])
			if err != nil {
				return nil, err
			}
		} else {
			buf = []byte(data)
		}

		if _, err := conn.Write(buf); err != nil {
			return nil, err
		}
		logWrite(fmt.Sprintf("Data (length:%d)", len(buf)))
	}

	return nil, nil
}

func (conn *Conn) sendDataFrame(param []byte) (interface{}, error) {
	var (
		p    SendDataFrameParam
		data []byte
	)

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	if p.DataLength > 0 || p.FillMaxFrameSize {
		dataLength := p.DataLength
		if p.FillMaxFrameSize {
			dataLength = conn.MaxFrameSize()
			if p.PadLength > 0 {
				dataLength -= 1
			}
		}
		data = randomData(dataLength)
	} else {
		data = []byte(p.Data)
	}

	defer conn.logWriteFrame()
	if p.PadLength > 0 {
		pad := randomData(uint32(p.PadLength))
		return nil, conn.framer.WriteDataPadded(p.StreamID, p.EndStream, data, pad)
	} else {
		return nil, conn.framer.WriteData(p.StreamID, p.EndStream, data)
	}
}

func (conn *Conn) sendHeadersFrame(param []byte) (interface{}, error) {
	var (
		p      SendHeadersFrameParam
		fields []Field
	)

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	if p.NoDefaultFields {
		fields = p.HeaderFields
	} else {
		fields = conn.setDefaultHeaderFields(p.HeaderFields)
	}

	if p.FillMaxFrameSize {
		num := int(conn.MaxFrameSize()/conn.maxFieldValueLength) + 1
		for i := 0; i < num; i++ {
			name := fmt.Sprintf("random-%d", i)
			fields = append(fields, Field{
				Name:  name,
				Value: string(randomData(conn.maxFieldValueLength)),
			})
		}
	}

	hp := http2.HeadersFrameParam{
		StreamID:      p.StreamID,
		EndStream:     p.EndStream,
		EndHeaders:    p.EndHeaders,
		PadLength:     p.PadLength,
		BlockFragment: conn.encodeHeaderFields(fields),
	}

	if p.Priority != nil {
		hp.Priority = http2.PriorityParam{
			StreamDep: p.Priority.StreamDependency,
			Exclusive: p.Priority.Exclusive,
			Weight:    p.Priority.Weight,
		}
	}

	defer conn.logWriteFrame()
	return nil, conn.framer.WriteHeaders(hp)
}

func (conn *Conn) sendPriorityFrame(param []byte) (interface{}, error) {
	var p SendPriorityFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	pp := http2.PriorityParam{
		StreamDep: p.StreamDependency,
		Exclusive: p.Exclusive,
		Weight:    p.Weight,
	}

	defer conn.logWriteFrame()
	return nil, conn.framer.WritePriority(p.StreamID, pp)
}

func (conn *Conn) sendRSTStreamFrame(param []byte) (interface{}, error) {
	var p SendRSTStreamFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	defer conn.logWriteFrame()
	return nil, conn.framer.WriteRSTStream(p.StreamID, errorCode[p.ErrorCode])
}

func (conn *Conn) sendSettingsFrame(param []byte) (interface{}, error) {
	var (
		p        SendSettingsFrameParam
		settings []http2.Setting
	)

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	if p.Ack {
		return nil, conn.framer.WriteSettingsAck()
	}

	for _, setting := range p.Settings {
		settings = append(settings, http2.Setting{
			ID:  settingID[setting.ID],
			Val: setting.Value,
		})
	}

	defer conn.logWriteFrame()
	return nil, conn.framer.WriteSettings(settings...)
}

func (conn *Conn) sendPingFrame(param []byte) (interface{}, error) {
	var (
		p    SendPingFrameParam
		data [8]byte
	)

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	copy(data[:], p.Data)

	defer conn.logWriteFrame()
	return nil, conn.framer.WritePing(p.Ack, data)
}

func (conn *Conn) sendGoAwayFrame(param []byte) (interface{}, error) {
	var p SendGoAwayFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	defer conn.logWriteFrame()
	return nil, conn.framer.WriteGoAway(p.LastStreamID, errorCode[p.ErrorCode], []byte(p.AdditionalDebugData))
}

func (conn *Conn) sendWindowUpdateFrame(param []byte) (interface{}, error) {
	var p SendWindowUpdateFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	defer conn.logWriteFrame()
	return nil, conn.framer.WriteWindowUpdate(p.StreamID, p.WindowSizeIncrement)
}

func (conn *Conn) sendContinuationFrame(param []byte) (interface{}, error) {
	var p SendContinuationFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	defer conn.logWriteFrame()
	return nil, conn.framer.WriteContinuation(p.StreamID, p.EndHeaders, conn.encodeHeaderFields(p.HeaderFields))
}

func (conn *Conn) waitHeadersFrame(param []byte) (interface{}, error) {
	var p WaitHeadersFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for {
		f, err := conn.readFrame()
		if err != nil {
			return nil, action.HandleConnectionFailure(err)
		}

		hf, ok := f.(*http2.HeadersFrame)
		if !ok {
			continue
		}

		matched := true

		if p.StreamID != 0 && hf.StreamID != p.StreamID {
			matched = false
		}

		if matched {
			return nil, nil
		}
	}
}

func (conn *Conn) waitRSTStreamFrame(param []byte) (interface{}, error) {
	var p WaitRSTStreamFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for {
		f, err := conn.readFrame()
		if err != nil {
			return nil, action.HandleConnectionFailure(err)
		}

		rsf, ok := f.(*http2.RSTStreamFrame)
		if !ok {
			continue
		}

		if p.StreamID != 0 && rsf.StreamID != p.StreamID {
			return nil, action.Failf("unexpected stream ID: %d", rsf.StreamID)
		}

		matched := false
		for _, ec := range p.ErrorCode {
			code := errorCode[ec]
			if rsf.ErrCode == code {
				matched = true
				break
			}
		}

		if !matched {
			return nil, action.Failf("unexpected error code: %s", rsf.ErrCode)
		}

		return nil, nil
	}
}

func (conn *Conn) waitSettingsFrame(param []byte) (interface{}, error) {
	var p WaitSettingsFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for {
		f, err := conn.readFrame()
		if err != nil {
			return nil, action.HandleConnectionFailure(err)
		}

		sf, ok := f.(*http2.SettingsFrame)
		if !ok {
			continue
		}

		if p.Ack {
			if sf.IsAck() != p.Ack {
				continue
			}

			return nil, nil
		}

		if sf.NumSettings() != len(p.Settings) {
			continue
		}

		matched := true
		for key, val := range p.Settings {
			id, ok := settingID[key]
			if !ok {
				matched = false
				break
			}

			v, ok := sf.Value(id)
			if !ok {
				matched = false
				break
			}

			if val != v {
				matched = false
				break
			}
		}

		if matched {
			return nil, nil
		}
	}
}

func (conn *Conn) waitPingFrame(param []byte) (interface{}, error) {
	var p WaitPingFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for {
		var valid = false

		f, err := conn.readFrame()
		if err != nil {
			return nil, action.HandleConnectionFailure(err)
		}

		pf, ok := f.(*http2.PingFrame)
		if !ok {
			continue
		}

		if pf.IsAck() == p.Ack {
			valid = true
		}

		if len(p.Data) > 0 {
			var data [8]byte
			copy(data[:], p.Data)

			if pf.Data != data {
				valid = false
			}
		}

		if valid {
			return nil, nil
		}
	}
}

func (conn *Conn) waitGoAwayFrame(param []byte) (interface{}, error) {
	var p WaitGoAwayFrameParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for {
		f, err := conn.readFrame()
		if err != nil {
			return nil, action.HandleConnectionFailure(err)
		}

		gaf, ok := f.(*http2.GoAwayFrame)
		if !ok {
			continue
		}

		if p.LastStreamID != 0 && gaf.LastStreamID != p.LastStreamID {
			return nil, action.Failf("unexpected last stream ID: %d", gaf.LastStreamID)
		}

		if len(p.DebugData) > 0 && p.DebugData != string(gaf.DebugData()) {
			return nil, action.Failf("unexpected debug data: %s", gaf.DebugData())
		}

		matched := false
		for _, ec := range p.ErrorCode {
			code := errorCode[ec]
			if gaf.ErrCode == code {
				matched = true
				break
			}
		}

		if !matched {
			return nil, action.Failf("unexpected error code: %s", gaf.ErrCode)
		}

		return nil, nil
	}
}

func (conn *Conn) waitConnectionError(param []byte) (interface{}, error) {
	var p WaitConnectionErrorParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for {
		f, err := conn.readFrame()
		if err != nil {
			if action.IsConnectionClosed(err) {
				return nil, nil
			}

			if action.IsTimeout(err) {
				return nil, action.ErrTimeout
			}

			return nil, err
		}

		gaf, ok := f.(*http2.GoAwayFrame)
		if !ok {
			continue
		}

		matched := false
		for _, code := range p.ErrorCode {
			if gaf.ErrCode == errorCode[code] {
				matched = true
				break
			}
		}

		if !matched {
			return nil, action.Failf("unexpected error code: %s", gaf.ErrCode)
		}

		return nil, nil
	}
}

func (conn *Conn) waitConnectionClose(param []byte) (interface{}, error) {
	for {
		_, err := conn.readFrame()
		if err != nil {
			if action.IsConnectionClosed(err) {
				return nil, nil
			}

			if action.IsTimeout(err) {
				return nil, action.ErrTimeout
			}

			return nil, err
		}
	}
}

func (conn *Conn) waitStreamError(param []byte) (interface{}, error) {
	var p WaitStreamErrorParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for {
		var errCode http2.ErrCode

		f, err := conn.readFrame()
		if err != nil {
			if action.IsConnectionClosed(err) {
				return nil, nil
			}

			if action.IsTimeout(err) {
				return nil, action.ErrTimeout
			}

			return nil, err
		}

		switch frame := f.(type) {
		case *http2.RSTStreamFrame:
			if p.StreamID != 0 && p.StreamID != f.Header().StreamID {
				continue
			}

			errCode = frame.ErrCode
		case *http2.GoAwayFrame:
			errCode = frame.ErrCode
		default:
			continue
		}

		matched := false
		for _, code := range p.ErrorCode {
			if errCode == errorCode[code] {
				matched = true
				break
			}
		}

		if !matched {
			return nil, action.Failf("unexpected error code: %s", errCode)
		}

		return nil, nil
	}
}

func (conn *Conn) waitStreamClose(param []byte) (interface{}, error) {
	var p WaitStreamCloseParam

	if err := json.Unmarshal(param, &p); err != nil {
		return nil, err
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	for {
		f, err := conn.readFrame()
		if err != nil {
			if action.IsConnectionClosed(err) {
				return nil, action.ErrConnectionClosed
			}

			if action.IsTimeout(err) {
				return nil, action.ErrTimeout
			}

			return nil, err
		}

		if f.Header().StreamID != p.StreamID {
			continue
		}

		switch frame := f.(type) {
		case *http2.DataFrame:
			if frame.StreamEnded() {
				return nil, nil
			}
		case *http2.HeadersFrame:
			if frame.StreamEnded() {
				return nil, nil
			}
		case *http2.RSTStreamFrame:
			if frame.ErrCode == http2.ErrCodeNo {
				return nil, nil
			}
		default:
			continue
		}

		return nil, nil
	}
}

func (conn *Conn) encodeHeaderFields(fields []Field) []byte {
	conn.encoderBuf.Reset()

	for _, field := range fields {
		conn.encoder.WriteField(hpack.HeaderField(field))
	}

	buf := make([]byte, conn.encoderBuf.Len())
	copy(buf, conn.encoderBuf.Bytes())

	return buf
}

func (conn *Conn) setDefaultHeaderFields(fields []Field) []Field {
	var (
		method    bool
		scheme    bool
		authority bool
		path      bool
	)

	for _, field := range fields {
		if !method && field.Name == ":method" {
			method = true
		}
		if !scheme && field.Name == ":scheme" {
			scheme = true
		}
		if !authority && field.Name == ":authority" {
			authority = true
		}
		if !path && field.Name == ":path" {
			path = true
		}
	}

	if !method {
		fields = append([]Field{{Name: ":method", Value: "GET"}}, fields...)
	}
	if !scheme {
		fields = append([]Field{{Name: ":scheme", Value: "http"}}, fields...)
	}
	if !authority {
		addr := conn.RemoteAddr()
		fields = append([]Field{{Name: ":authority", Value: addr.String()}}, fields...)
	}
	if !path {
		fields = append([]Field{{Name: ":path", Value: "/"}}, fields...)
	}

	return fields
}

func (conn *Conn) logWriteFrame() {
	f, _ := conn.debugFramer.ReadFrame()
	if f != nil {
		logWrite(getEvent(f))
	}
}

func randomData(l uint32) []byte {
	data := make([]byte, l)

	_, err := rand.Read(data)
	if err != nil {
		panic(err)
	}

	return data
}

func logWrite(msg string) {
	log.Debug(fmt.Sprintf("send: %s", msg))
}

func getEvent(f http2.Frame) string {
	return fmt.Sprintf(
		"%s Frame (length:%d, flags:0x%02x, stream_id:%d)",
		f.Header().Type, f.Header().Length, f.Header().Flags, f.Header().StreamID,
	)
}
