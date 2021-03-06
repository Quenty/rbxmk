package rbxmk

import (
	"errors"
	"fmt"
	"io"
	"sort"
)

// EOD indicates that a Data cannot be drilled into.
var EOD = errors.New("end of drill")

type Data interface {
	// Type returns a string representation of the Data's type.
	Type() string

	// Drill drills into the Data using inref, returning another Data that
	// represents the result. It also returns the reference after it has been
	// processed. In case of an error, the original Data and inref is
	// returned. If inref is empty, or the Data cannot be drilled into, then
	// an EOD error is returned.
	Drill(opt Options, inref []string) (outdata Data, outref []string, err error)

	// Merge is used to merge input Data into output Data. Three kinds of Data
	// are received. rootdata is the top-level data returned by an output
	// scheme handler. drilldata is a value within the rootdata, which was
	// selected by one or more Drills. If no drills were used, then drilldata
	// will be equal to rootdata. Depending on the scheme, both rootdata and
	// drilldata may be nil. The receiver is the data to be merged. Merge
	// returns the resulting Data after it has been merged.
	Merge(opt Options, rootdata, drilldata Data) (outdata Data, err error)
}

type DataTypeError struct {
	dataName string
}

func (err DataTypeError) Error() string {
	return fmt.Sprintf("unexpected Data type: %s", err.dataName)
}

func NewDataTypeError(data Data) error {
	if data == nil {
		return DataTypeError{dataName: "nil"}
	}
	return DataTypeError{dataName: data.Type()}
}

type MergeError struct {
	indata, drilldata string
	msg               error
}

func (err *MergeError) Error() string {
	if err.msg == nil {
		return fmt.Sprintf("cannot merge %s into %s", err.indata, err.drilldata)
	}
	return fmt.Sprintf("cannot merge %s into %s: %s", err.indata, err.drilldata, err.msg.Error())
}

func NewMergeError(indata, drilldata Data, msg error) error {
	err := &MergeError{"nil", "nil", msg}
	if indata != nil {
		err.indata = indata.Type()
	}
	if drilldata != nil {
		err.drilldata = drilldata.Type()
	}
	return err
}

type Format struct {
	Name  string
	Ext   string
	Codec InitFormatCodec
	// CanEncode    func(data Data) bool
}

type FormatDecoder interface {
	Decode(r io.Reader, data *Data) (err error)
}

type FormatEncoder interface {
	Encode(w io.Writer, data Data) (err error)
}

type FormatCodec interface {
	FormatDecoder
	FormatEncoder
}

type InitFormatCodec func(opt Options, ctx interface{}) (codec FormatCodec)

type Formats struct {
	f map[string]*Format
}

func NewFormats() *Formats {
	return &Formats{f: map[string]*Format{}}
}

func (fs *Formats) Register(formats ...Format) error {
	for i, f := range formats {
		if f.Ext == "" {
			return fmt.Errorf("format #%d must have non-empty Ext", i)
		}
		if _, registered := fs.f[f.Ext]; registered {
			return fmt.Errorf("format \"%s\" is already registered", f.Ext)
		}
		if f.Codec == nil {
			return fmt.Errorf("format \"%s\" must have Codec function", f.Ext)
		}
	}
	for _, f := range formats {
		format := f
		fs.f[format.Ext] = &format
	}
	return nil
}

func (fs *Formats) List() []Format {
	l := make([]Format, len(fs.f))
	i := 0
	for _, f := range fs.f {
		l[i] = *f
		i++
	}
	sort.Slice(l, func(i, j int) bool {
		return l[i].Ext < l[j].Ext
	})
	return l
}

func (fs *Formats) Registered(ext string) (registered bool) {
	_, registered = fs.f[ext]
	return registered
}

func (fs *Formats) Name(ext string) (name string) {
	f, registered := fs.f[ext]
	if !registered {
		return ""
	}
	return f.Name
}

func (fs *Formats) Decoder(ext string, opt Options, ctx interface{}) (dec FormatDecoder) {
	f, registered := fs.f[ext]
	if !registered {
		return nil
	}
	return f.Codec(opt, ctx)
}

func (fs *Formats) Decode(ext string, opt Options, ctx interface{}, r io.Reader, data *Data) (err error) {
	f, registered := fs.f[ext]
	if !registered {
		return nil
	}
	return f.Codec(opt, ctx).Decode(r, data)
}

func (fs *Formats) Encoder(ext string, opt Options, ctx interface{}) (enc FormatEncoder) {
	f, registered := fs.f[ext]
	if !registered {
		return nil
	}
	return f.Codec(opt, ctx)
}

func (fs *Formats) Encode(ext string, opt Options, ctx interface{}, w io.Writer, data Data) (err error) {
	f, registered := fs.f[ext]
	if !registered {
		return nil
	}
	return f.Codec(opt, ctx).Encode(w, data)
}
