//go:generate gondn_tlv_gen
package gen_signature

import (
	"crypto/sha256"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

//+tlv-model:nocopy,private
type T1 struct {
	//+field:natural
	H1 uint64 `tlv:"0x01"`
	//+field:offsetMarker
	sigCoverStart enc.PlaceHolder
	//+field:natural:optional
	H2 *uint64 `tlv:"0x02"`
	//+field:wire
	C enc.Wire `tlv:"0x03"`
	//+field:signature:sigCoverStart:sigCovered
	Sig enc.Wire `tlv:"0x04"`

	//+field:procedureArgument:enc.Wire
	sigCovered enc.PlaceHolder
}

func (v *T1) Encode(estLen uint, value []byte) (enc.Wire, enc.Wire) {
	encoder := T1Encoder{
		Sig_estLen: estLen,
	}
	encoder.Init(v)
	wire := encoder.Encode(v)
	// Compute signature
	if encoder.Sig_wireIdx >= 0 {
		wire[encoder.Sig_wireIdx] = value
		// Fix length, assuming len(value) <= estLen < 253
		buf := wire[encoder.Sig_wireIdx-1]
		buf[len(buf)-1] = byte(len(value))
	}

	return wire, encoder.sigCovered
}

func ReadT1(reader enc.ParseReader) (*T1, enc.Wire, error) {
	context := T1ParsingContext{}
	context.Init()
	ret, err := context.Parse(reader, false)
	if err != nil {
		return nil, nil, err
	}
	return ret, context.sigCovered, nil
}

//+tlv-model:nocopy,private
type T2 struct {
	//+field:interestName:sigCovered
	Name enc.Name `tlv:"0x01"`
	//+field:offsetMarker
	sigCoverStart enc.PlaceHolder
	//+field:offsetMarker
	digestCoverStart enc.PlaceHolder
	//+field:wire
	C enc.Wire `tlv:"0x03"`
	//+field:signature:sigCoverStart:sigCovered
	Sig enc.Wire `tlv:"0x04"`

	//+field:offsetMarker
	digestCoverEnd enc.PlaceHolder
	//+field:procedureArgument:enc.Wire
	sigCovered enc.PlaceHolder
}

func (v *T2) Encode(estLen uint, value []byte, needDigest bool) (enc.Wire, enc.Wire) {
	encoder := T2Encoder{
		Sig_estLen:      estLen,
		Name_needDigest: needDigest,
	}
	encoder.Init(v)
	wire := encoder.Encode(v)
	// Compute signature
	if encoder.Sig_wireIdx >= 0 {
		wire[encoder.Sig_wireIdx] = value
		// Fix length, assuming len(value) <= estLen < 253
		buf := wire[encoder.Sig_wireIdx-1]
		buf[len(buf)-1] = byte(len(value))
	}
	// Compute digest
	if needDigest {
		buf := wire[encoder.Name_wireIdx]
		digestBuf := buf[encoder.Name_pos : encoder.Name_pos+32]

		digestCovered := enc.Wire(nil)
		if encoder.digestCoverStart_wireIdx == encoder.digestCoverEnd_wireIdx {
			buf := wire[encoder.digestCoverStart_wireIdx]
			coveredPart := buf[encoder.digestCoverStart_pos:encoder.digestCoverEnd_pos]
			digestCovered = enc.Wire{coveredPart}
		} else {
			coverStart := wire[encoder.digestCoverStart_wireIdx][encoder.digestCoverStart_pos:]
			digestCovered = append(digestCovered, coverStart)
			for i := encoder.digestCoverStart_wireIdx + 1; i < encoder.digestCoverEnd_wireIdx; i++ {
				digestCovered = append(digestCovered, wire[i])
			}
			if encoder.digestCoverEnd_pos > 0 { // Actually always false
				coverEnd := wire[encoder.digestCoverEnd_wireIdx][:encoder.digestCoverEnd_pos]
				digestCovered = append(digestCovered, coverEnd)
			}
		}

		h := sha256.New()
		for _, buf := range digestCovered {
			_, err := h.Write(buf)
			if err != nil {
				return nil, nil
			}
		}
		copy(digestBuf, h.Sum(nil))
	}

	return wire, encoder.sigCovered
}
