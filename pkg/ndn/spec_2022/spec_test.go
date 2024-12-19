package spec_2022_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	basic_engine "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	"github.com/zjkmxy/go-ndn/pkg/ndn"
	"github.com/zjkmxy/go-ndn/pkg/ndn/spec_2022"
	"github.com/zjkmxy/go-ndn/pkg/security"
	"github.com/zjkmxy/go-ndn/pkg/utils"
)

func TestMakeDataBasic(t *testing.T) {
	utils.SetTestingT(t)

	spec := spec_2022.Spec{}

	data, err := spec.MakeData(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.DataConfig{
			ContentType: utils.IdPtr(ndn.ContentTypeBlob),
		},
		nil,
		security.NewSha256Signer(),
	)
	require.NoError(t, err)
	require.Equal(t, []byte(
		"\x06\x42\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix"+
			"\x14\x03\x18\x01\x00"+
			"\x16\x03\x1b\x01\x00"+
			"\x17 \x7f1\xe4\t\xc5z/\x1d\r\xdaVh8\xfd\xd9\x94"+
			"\xd8'S\x13[\xd7\x15\xa5\x9d%^\x80\xf2\xab\xf0\xb5"),
		data.Wire.Join())

	data, err = spec.MakeData(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.DataConfig{
			ContentType: utils.IdPtr(ndn.ContentTypeBlob),
		},
		enc.Wire{[]byte("01020304")},
		security.NewSha256Signer(),
	)
	require.NoError(t, err)
	require.Equal(t, []byte(
		"\x06L\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix"+
			"\x14\x03\x18\x01\x00"+
			"\x15\x0801020304"+
			"\x16\x03\x1b\x01\x00"+
			"\x17 \x94\xe9\xda\x91\x1a\x11\xfft\x02i:G\x0cO\xdd!"+
			"\xe0\xc7\xb6\xfd\x8f\x9cn\xc5\x93{\x93\x04\xe0\xdf\xa6S"),
		data.Wire.Join())

	data, err = spec.MakeData(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.DataConfig{
			ContentType: utils.IdPtr(ndn.ContentTypeBlob),
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, []byte(
		"\x06\x1b\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix"+
			"\x14\x03\x18\x01\x00"),
		data.Wire.Join())

	data, err = spec.MakeData(
		utils.WithoutErr(enc.NameFromStr("/E")),
		&ndn.DataConfig{
			ContentType: nil,
		},
		enc.Wire{},
		security.NewSha256Signer(),
	)
	require.NoError(t, err)
	require.Equal(t, utils.WithoutErr(hex.DecodeString(
		"06300703080145"+
			"1400150016031b0100"+
			"1720f965ee682c6973c3cbaa7b69e4c7063680f83be93a46be2ccc98686134354b66")),
		data.Wire.Join())
}

func TestMakeDataMetaInfo(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	data, err := spec.MakeData(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix/37=%00")),
		&ndn.DataConfig{
			ContentType:  utils.IdPtr(ndn.ContentTypeBlob),
			Freshness:    utils.IdPtr(1000 * time.Millisecond),
			FinalBlockID: utils.IdPtr(enc.NewSequenceNumComponent(2)),
		},
		nil,
		security.NewSha256Signer(),
	)
	require.NoError(t, err)
	require.Equal(t, []byte(
		"\x06\x4e\x07\x17\x08\x05local\x08\x03ndn\x08\x06prefix\x25\x01\x00"+
			"\x14\x0c\x18\x01\x00\x19\x02\x03\xe8\x1a\x03\x3a\x01\x02"+
			"\x16\x03\x1b\x01\x00"+
			"\x17 \x0f^\xa1\x0c\xa7\xf5Fb\xf0\x9cOT\xe0FeC\x8f92\x04\x9d\xabP\x80o'\x94\xaa={hQ"),
		data.Wire.Join())
}

type testSigner struct{}

func (testSigner) SigInfo() (*ndn.SigConfig, error) {
	name, _ := enc.NameFromStr("/KEY")
	return &ndn.SigConfig{
		Type:    ndn.SigType(200),
		KeyName: name,
	}, nil
}

func (testSigner) EstimateSize() uint {
	return 10
}

func (testSigner) ComputeSigValue(enc.Wire) ([]byte, error) {
	return []byte{0, 0, 0, 0, 0}, nil
}

func TestMakeDataShrink(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	data, err := spec.MakeData(
		utils.WithoutErr(enc.NameFromStr("/test")),
		&ndn.DataConfig{
			ContentType: utils.IdPtr(ndn.ContentTypeBlob),
		},
		nil,
		testSigner{},
	)
	require.NoError(t, err)
	require.Equal(t, []byte{
		0x6, 0x22, 0x7, 0x6, 0x8, 0x4, 0x74, 0x65, 0x73, 0x74, 0x14, 0x3, 0x18, 0x1, 0x0,
		0x16, 0xc, 0x1b, 0x1, 0xc8, 0x1c, 0x7, 0x7, 0x5, 0x8, 0x3, 0x4b, 0x45, 0x59,
		0x17, 0x5, 0x0, 0x0, 0x0, 0x0, 0x0},
		data.Wire.Join())
}

func TestReadDataBasic(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	data, covered, err := spec.ReadData(enc.NewBufferReader([]byte(
		"\x06\x42\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix" +
			"\x14\x03\x18\x01\x00" +
			"\x16\x03\x1b\x01\x00" +
			"\x17 \x7f1\xe4\t\xc5z/\x1d\r\xdaVh8\xfd\xd9\x94" +
			"\xd8'S\x13[\xd7\x15\xa5\x9d%^\x80\xf2\xab\xf0\xb5"),
	))
	require.NoError(t, err)
	require.Equal(t, "/local/ndn/prefix", data.Name().String())
	require.Equal(t, ndn.ContentTypeBlob, *data.ContentType())
	require.True(t, data.Freshness() == nil)
	require.True(t, data.FinalBlockID() == nil)
	require.Equal(t, ndn.SignatureDigestSha256, data.Signature().SigType())
	require.True(t, data.Content() == nil)
	h := sha256.New()
	for _, c := range covered {
		h.Write(c)
	}
	sig := h.Sum(nil)
	require.Equal(t, sig, data.Signature().SigValue())

	data, covered, err = spec.ReadData(enc.NewBufferReader([]byte(
		"\x06L\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix" +
			"\x14\x03\x18\x01\x00" +
			"\x15\x0801020304" +
			"\x16\x03\x1b\x01\x00" +
			"\x17 \x94\xe9\xda\x91\x1a\x11\xfft\x02i:G\x0cO\xdd!" +
			"\xe0\xc7\xb6\xfd\x8f\x9cn\xc5\x93{\x93\x04\xe0\xdf\xa6S"),
	))
	require.NoError(t, err)
	require.Equal(t, "/local/ndn/prefix", data.Name().String())
	require.Equal(t, ndn.ContentTypeBlob, *data.ContentType())
	require.True(t, data.Freshness() == nil)
	require.True(t, data.FinalBlockID() == nil)
	require.Equal(t, ndn.SignatureDigestSha256, data.Signature().SigType())
	require.Equal(t, []byte("01020304"), data.Content().Join())
	h = sha256.New()
	for _, c := range covered {
		h.Write(c)
	}
	sig = h.Sum(nil)
	require.Equal(t, sig, data.Signature().SigValue())

	data, _, err = spec.ReadData(enc.NewBufferReader([]byte(
		"\x06\x1b\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix" +
			"\x14\x03\x18\x01\x00"),
	))
	require.NoError(t, err)
	require.Equal(t, "/local/ndn/prefix", data.Name().String())
	require.Equal(t, ndn.ContentTypeBlob, *data.ContentType())
	require.True(t, data.Freshness() == nil)
	require.True(t, data.FinalBlockID() == nil)
	require.Equal(t, ndn.SignatureNone, data.Signature().SigType())
	require.True(t, data.Content() == nil)
	require.True(t, data.Signature().SigValue() == nil)

	data, covered, err = spec.ReadData(enc.NewBufferReader(utils.WithoutErr(hex.DecodeString(
		"06300703080145" +
			"1400150016031b0100" +
			"1720f965ee682c6973c3cbaa7b69e4c7063680f83be93a46be2ccc98686134354b66"),
	)))
	require.NoError(t, err)
	require.Equal(t, "/E", data.Name().String())
	require.True(t, data.ContentType() == nil)
	require.True(t, data.Freshness() == nil)
	require.True(t, data.FinalBlockID() == nil)
	require.Equal(t, ndn.SignatureDigestSha256, data.Signature().SigType())
	require.Equal(t, 0, len(data.Content().Join()))
	h = sha256.New()
	for _, c := range covered {
		h.Write(c)
	}
	sig = h.Sum(nil)
	require.Equal(t, sig, data.Signature().SigValue())
}

func TestReadDataMetaInfo(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	data, covered, err := spec.ReadData(enc.NewBufferReader([]byte(
		"\x06\x4e\x07\x17\x08\x05local\x08\x03ndn\x08\x06prefix\x25\x01\x00" +
			"\x14\x0c\x18\x01\x00\x19\x02\x03\xe8\x1a\x03\x3a\x01\x02" +
			"\x16\x03\x1b\x01\x00" +
			"\x17 \x0f^\xa1\x0c\xa7\xf5Fb\xf0\x9cOT\xe0FeC\x8f92\x04\x9d\xabP\x80o'\x94\xaa={hQ"),
	))
	require.NoError(t, err)
	require.Equal(t, "/local/ndn/prefix/37=%00", data.Name().String())
	require.Equal(t, ndn.ContentTypeBlob, *data.ContentType())
	require.Equal(t, 1000*time.Millisecond, *data.Freshness())
	require.Equal(t, enc.NewSequenceNumComponent(2), *data.FinalBlockID())
	require.Equal(t, ndn.SignatureDigestSha256, data.Signature().SigType())
	require.True(t, data.Content() == nil)
	h := sha256.New()
	for _, c := range covered {
		h.Write(c)
	}
	sig := h.Sum(nil)
	require.Equal(t, sig, data.Signature().SigValue())
}

func TestMakeIntBasic(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	interest, err := spec.MakeInterest(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.InterestConfig{
			Lifetime: utils.IdPtr(4 * time.Second),
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, "/local/ndn/prefix", interest.FinalName.String())
	require.Equal(t, []byte("\x05\x1a\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix\x0c\x02\x0f\xa0"),
		interest.Wire.Join())

	interest, err = spec.MakeInterest(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.InterestConfig{
			CanBePrefix: true,
			MustBeFresh: true,
			Lifetime:    utils.IdPtr(10 * time.Millisecond),
			HopLimit:    utils.IdPtr[uint](1),
			Nonce:       utils.IdPtr[uint64](0),
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, []byte(
		"\x05\x26\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix"+
			"\x21\x00\x12\x00\x0a\x04\x00\x00\x00\x00\x0c\x01\x0a\x22\x01\x01"),
		interest.Wire.Join())

	interest, err = spec.MakeInterest(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.InterestConfig{
			Lifetime: utils.IdPtr(4 * time.Second),
			Nonce:    utils.IdPtr[uint64](0x01020304),
			ForwardingHint: []enc.Name{
				utils.WithoutErr(enc.NameFromStr("/name/A")),
				utils.WithoutErr(enc.NameFromStr("/ndn/B")),
				utils.WithoutErr(enc.NameFromBytes([]byte("\x07\x0d\x08\x0bshekkuenseu"))),
			},
		},
		nil,
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, []byte(
		"\x05\x46\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix\x1e\x24"+
			"\x07\x09\x08\x04name\x08\x01A"+
			"\x07\x08\x08\x03ndn\x08\x01B"+
			"\x07\r\x08\x0bshekkuenseu"+
			"\x0a\x04\x01\x02\x03\x04\x0c\x02\x0f\xa0"),
		interest.Wire.Join())
}

func TestMakeIntLargeAppParam(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	appParam := make([]byte, 384)
	for i := range appParam {
		appParam[i] = byte(i & 0xff)
	}
	encoded, err := spec.MakeInterest(
		utils.WithoutErr(enc.NameFromStr("/interest/with/large/prefix")),
		&ndn.InterestConfig{
			Lifetime: utils.IdPtr(4 * time.Second),
		},
		enc.Wire{appParam},
		security.NewHmacIntSigner([]byte("temp-hmac-key"), basic_engine.NewTimer()),
	)
	require.NoError(t, err)

	interest, _, err := spec.ReadInterest(enc.NewWireReader(encoded.Wire))
	require.NoError(t, err)
	require.Equal(t, appParam, interest.AppParam().Join())
	require.True(t, interest.Name().Equal(encoded.FinalName))
}

func TestMakeIntSign(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	interest, err := spec.MakeInterest(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.InterestConfig{
			Lifetime: utils.IdPtr(4 * time.Second),
		},
		enc.Wire{[]byte{1, 2, 3, 4}},
		nil,
	)
	require.NoError(t, err)
	require.Equal(t,
		"/local/ndn/prefix/params-sha256=47756f21fe0ee265149aa2be3c63c538a72378e9b0a58b39c5916367d35bda10",
		interest.FinalName.String())
	require.Equal(t, []byte(
		"\x05\x42\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix"+
			"\x02 \x47\x75\x6f\x21\xfe\x0e\xe2\x65\x14\x9a\xa2\xbe\x3c\x63\xc5\x38"+
			"\xa7\x23\x78\xe9\xb0\xa5\x8b\x39\xc5\x91\x63\x67\xd3\x5b\xda\x10"+
			"\x0c\x02\x0f\xa0\x24\x04\x01\x02\x03\x04"),
		interest.Wire.Join())

	// "/test/params-sha256=FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF/ndn" is not supported yet

	interest, err = spec.MakeInterest(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.InterestConfig{
			Lifetime: utils.IdPtr(4 * time.Second),
			Nonce:    utils.IdPtr[uint64](0x6c211166),
		},
		enc.Wire{[]byte{1, 2, 3, 4}},
		security.NewSha256Signer(),
	)
	require.NoError(t, err)
	require.Equal(t,
		"/local/ndn/prefix/params-sha256=8e6e36d7eabcde43756140c90bda09d500d2a577f2f533b569f0441df0a7f9e2",
		interest.FinalName.String())
	require.Equal(t, []byte(
		"\x05\x6f\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix"+
			"\x02 \x8e\x6e\x36\xd7\xea\xbc\xde\x43\x75\x61\x40\xc9\x0b\xda\x09\xd5"+
			"\x00\xd2\xa5\x77\xf2\xf5\x33\xb5\x69\xf0\x44\x1d\xf0\xa7\xf9\xe2"+
			"\x0a\x04\x6c\x21\x11\x66\x0c\x02\x0f\xa0"+
			"\x24\x04\x01\x02\x03\x04"+
			"\x2c\x03\x1b\x01\x00"+
			"\x2e \xea\xa8\xf0\x99\x08\x63\x78\x95\x1d\xe0\x5f\xf1\xde\xbb\xc1\x18"+
			"\xb5\x21\x8b\x2f\xca\xa0\xb5\x1d\x18\xfa\xbc\x29\xf5\x4d\x58\xff"),
		interest.Wire.Join())

	interest, err = spec.MakeInterest(
		utils.WithoutErr(enc.NameFromStr("/local/ndn/prefix")),
		&ndn.InterestConfig{
			Lifetime: utils.IdPtr(4 * time.Second),
			Nonce:    utils.IdPtr[uint64](0x6c211166),
		},
		enc.Wire{},
		security.NewSha256Signer(),
	)
	require.NoError(t, err)
	require.Equal(t,
		"/local/ndn/prefix/params-sha256=4077a57049d83848b525a423ab978e6480f96d5ca38a80a5e2d6e250a617be4f",
		interest.FinalName.String())
	require.Equal(t, []byte(
		"\x05\x6b\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix"+
			"\x02 \x40\x77\xa5\x70\x49\xd8\x38\x48\xb5\x25\xa4\x23\xab\x97\x8e\x64"+
			"\x80\xf9\x6d\x5c\xa3\x8a\x80\xa5\xe2\xd6\xe2\x50\xa6\x17\xbe\x4f"+
			"\x0a\x04\x6c\x21\x11\x66\x0c\x02\x0f\xa0"+
			"\x24\x00"+
			"\x2c\x03\x1b\x01\x00"+
			"\x2e \x09\x4e\x00\x9d\x74\x59\x82\x5c\xa0\x2d\xaa\xb7\xad\x60\x48\x30"+
			"\x39\x19\xd8\x99\x80\x25\xbe\xff\xa6\xf9\x96\x79\xd6\x5e\x9f\x62"),
		interest.Wire.Join())
}

func TestReadIntBasic(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	interest, _, err := spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x05\x1a\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix\x0c\x02\x0f\xa0"),
	))
	require.NoError(t, err)
	require.Equal(t, "/local/ndn/prefix", interest.Name().String())
	require.Equal(t, 4*time.Second, *interest.Lifetime())
	require.True(t, interest.AppParam() == nil)
	require.False(t, interest.CanBePrefix())
	require.False(t, interest.MustBeFresh())
	require.True(t, interest.Nonce() == nil)
	require.True(t, interest.HopLimit() == nil)
	require.True(t, interest.Signature().SigType() == ndn.SignatureNone)

	interest, _, err = spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x05\x26\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix" +
			"\x21\x00\x12\x00\x0a\x04\x00\x00\x00\x00\x0c\x01\x0a\x22\x01\x01"),
	))
	require.NoError(t, err)
	require.Equal(t, "/local/ndn/prefix", interest.Name().String())
	require.Equal(t, 10*time.Millisecond, *interest.Lifetime())
	require.True(t, interest.AppParam() == nil)
	require.True(t, interest.CanBePrefix())
	require.True(t, interest.MustBeFresh())
	require.Equal(t, uint64(0), *interest.Nonce())
	require.Equal(t, uint(1), *interest.HopLimit())
	require.True(t, interest.Signature().SigType() == ndn.SignatureNone)

	interest, _, err = spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x05\x42\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix" +
			"\x02 \x47\x75\x6f\x21\xfe\x0e\xe2\x65\x14\x9a\xa2\xbe\x3c\x63\xc5\x38" +
			"\xa7\x23\x78\xe9\xb0\xa5\x8b\x39\xc5\x91\x63\x67\xd3\x5b\xda\x10" +
			"\x0c\x02\x0f\xa0\x24\x04\x01\x02\x03\x04"),
	))
	require.NoError(t, err)
	require.Equal(t,
		"/local/ndn/prefix/params-sha256=47756f21fe0ee265149aa2be3c63c538a72378e9b0a58b39c5916367d35bda10",
		interest.Name().String())
	require.Equal(t, 4*time.Second, *interest.Lifetime())
	require.False(t, interest.CanBePrefix())
	require.False(t, interest.MustBeFresh())
	require.Equal(t, []byte{1, 2, 3, 4}, interest.AppParam().Join())
	require.True(t, interest.Signature().SigType() == ndn.SignatureNone)

	// Reject wrong digest
	_, _, err = spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x05\x42\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix" +
			"\x02 \x47\x75\x6f\x21\xfe\x0e\xe2\x65\x14\x9a\xa2\xbe\x3c\x63\xc5\x38" +
			"\xa7\x23\x78\xe9\xb0\xa5\x8b\x39\xc5\x91\x63\x67\xd3\x5b\x00\x00" +
			"\x0c\x02\x0f\xa0\x24\x04\x01\x02\x03\x04"),
	))
	require.Error(t, err)

	var covered enc.Wire
	interest, covered, err = spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x05\x6f\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix" +
			"\x02 \x8e\x6e\x36\xd7\xea\xbc\xde\x43\x75\x61\x40\xc9\x0b\xda\x09\xd5" +
			"\x00\xd2\xa5\x77\xf2\xf5\x33\xb5\x69\xf0\x44\x1d\xf0\xa7\xf9\xe2" +
			"\x0a\x04\x6c\x21\x11\x66\x0c\x02\x0f\xa0" +
			"\x24\x04\x01\x02\x03\x04" +
			"\x2c\x03\x1b\x01\x00" +
			"\x2e \xea\xa8\xf0\x99\x08\x63\x78\x95\x1d\xe0\x5f\xf1\xde\xbb\xc1\x18" +
			"\xb5\x21\x8b\x2f\xca\xa0\xb5\x1d\x18\xfa\xbc\x29\xf5\x4d\x58\xff"),
	))
	require.NoError(t, err)
	require.Equal(t,
		"/local/ndn/prefix/params-sha256=8e6e36d7eabcde43756140c90bda09d500d2a577f2f533b569f0441df0a7f9e2",
		interest.Name().String())
	require.Equal(t, uint64(0x6c211166), *interest.Nonce())
	require.Equal(t, []byte{1, 2, 3, 4}, interest.AppParam().Join())
	require.True(t, interest.Signature().SigType() == ndn.SignatureDigestSha256)
	h := sha256.New()
	for _, c := range covered {
		h.Write(c)
	}
	sig := h.Sum(nil)
	require.Equal(t, sig, interest.Signature().SigValue())

	interest, covered, err = spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x05\x6b\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix" +
			"\x02 \x40\x77\xa5\x70\x49\xd8\x38\x48\xb5\x25\xa4\x23\xab\x97\x8e\x64" +
			"\x80\xf9\x6d\x5c\xa3\x8a\x80\xa5\xe2\xd6\xe2\x50\xa6\x17\xbe\x4f" +
			"\x0a\x04\x6c\x21\x11\x66\x0c\x02\x0f\xa0" +
			"\x24\x00" +
			"\x2c\x03\x1b\x01\x00" +
			"\x2e \x09\x4e\x00\x9d\x74\x59\x82\x5c\xa0\x2d\xaa\xb7\xad\x60\x48\x30" +
			"\x39\x19\xd8\x99\x80\x25\xbe\xff\xa6\xf9\x96\x79\xd6\x5e\x9f\x62"),
	))
	require.NoError(t, err)
	require.Equal(t,
		"/local/ndn/prefix/params-sha256=4077a57049d83848b525a423ab978e6480f96d5ca38a80a5e2d6e250a617be4f",
		interest.Name().String())
	require.Equal(t, uint64(0x6c211166), *interest.Nonce())
	require.Equal(t, []byte{}, interest.AppParam().Join())
	require.True(t, interest.Signature().SigType() == ndn.SignatureDigestSha256)
	h = sha256.New()
	for _, c := range covered {
		h.Write(c)
	}
	sig = h.Sum(nil)
	require.Equal(t, sig, interest.Signature().SigValue())
}

func TestReadIntErrors(t *testing.T) {
	utils.SetTestingT(t)
	spec := spec_2022.Spec{}

	_, _, err := spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x05\x6b\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix"),
	))
	require.Error(t, err)

	_, _, err = spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x05\x6b\x07\x14\x08\x05local\x08\x03ndn\x08\x06prefix"),
	))
	require.Error(t, err)

	_, _, err = spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x06\x6b\x07\x36\x08\x05local\x08\x03ndn\x08\x06prefix"),
	))
	require.Error(t, err)

	_, _, err = spec.ReadInterest(enc.NewBufferReader([]byte(
		"\x01\x00"),
	))
	require.Error(t, err)
}
