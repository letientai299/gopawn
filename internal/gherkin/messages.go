package gherkin

import (
	"bufio"
	"fmt"
	"gopawn/internal/msg"
	"io"
	"io/ioutil"
	"math"
	"strings"

	gio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

func Messagesx(
	paths []string,
	sourceStream io.Reader,
	language string,
	includeSource bool,
	includeGherkinDocument bool,
	includePickles bool,
	outStream io.Writer,
	json bool,
) ([]msg.Envelope, error) {
	var result []msg.Envelope
	var err error

	handleMessage := func(result []msg.Envelope, message *msg.Envelope) ([]msg.Envelope, error) {
		if outStream != nil {
			if json {
				ma := jsonpb.Marshaler{}
				msgJson, err := ma.MarshalToString(message)
				if err != nil {
					return result, err
				}
				out := bufio.NewWriter(outStream)
				out.WriteString(msgJson)
				out.WriteString("\n")
			} else {
				bytes, err := proto.Marshal(message)
				if err != nil {
					return result, err
				}
				outStream.Write(proto.EncodeVarint(uint64(len(bytes))))
				outStream.Write(bytes)
			}

		} else {
			result = append(result, *message)
		}

		return result, err
	}

	processSource := func(source *msg.Source) error {
		if includeSource {
			result, err = handleMessage(result, &msg.Envelope{
				Message: &msg.Envelope_Source{
					Source: source,
				},
			})
		}
		doc, err := ParseGherkinDocumentForLanguage(strings.NewReader(source.Data), language)
		if errs, ok := err.(parseErrors); ok {
			// expected parse errors
			for _, err := range errs {
				if pe, ok := err.(*parseError); ok {
					result, err = handleMessage(result, pe.asAttachment(source.Uri))
				} else {
					return fmt.Errorf("parse feature file: %s, unexpected error: %+v\n", source.Uri, err)
				}
			}
			return nil
		}

		if includeGherkinDocument {
			doc.Uri = source.Uri
			result, err = handleMessage(result, &msg.Envelope{
				Message: &msg.Envelope_GherkinDocument{
					GherkinDocument: doc,
				},
			})
		}

		if includePickles {
			for _, pickle := range Pickles(*doc, source.Uri, source.Data) {
				result, err = handleMessage(result, &msg.Envelope{
					Message: &msg.Envelope_Pickle{
						Pickle: pickle,
					},
				})
			}
		}
		return nil
	}

	if len(paths) == 0 {
		reader := gio.NewDelimitedReader(sourceStream, math.MaxInt32)
		for {
			wrapper := &msg.Envelope{}
			err := reader.ReadMsg(wrapper)
			if err == io.EOF {
				break
			}

			switch t := wrapper.Message.(type) {
			case *msg.Envelope_Source:
				processSource(t.Source)
			}
		}
	} else {
		for _, path := range paths {
			in, err := ioutil.ReadFile(path)
			if err != nil {
				return result, fmt.Errorf("read feature file: %s - %+v", path, err)
			}
			source := &msg.Source{
				Uri:  path,
				Data: string(in),
				Media: &msg.Media{
					Encoding:    msg.Media_UTF8,
					ContentType: "text/x.cucumber.gherkin+plain",
				},
			}
			processSource(source)
		}
	}

	return result, err
}

func (a *parseError) asAttachment(uri string) *msg.Envelope {
	return &msg.Envelope{
		Message: &msg.Envelope_Attachment{
			Attachment: &msg.Attachment{
				Data: a.Error(),
				Source: &msg.SourceReference{
					Uri: uri,
					Location: &msg.Location{
						Line:   uint32(a.loc.Line),
						Column: uint32(a.loc.Column),
					},
				},
			}},
	}
}
