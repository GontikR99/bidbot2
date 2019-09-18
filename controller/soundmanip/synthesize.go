package soundmanip

import (
	"bytes"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"context"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
	"os"
)

func Synthesize(credpath string, text string) (opusFile *OpusFile, err error) {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credpath)
	client, err := texttospeech.NewClient(context.Background())
	if err != nil {
		return
	}
	defer client.Close()
	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-GB",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
			Name:         "en-GB-Wavenet-A",
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_OGG_OPUS,
		},
	}
	resp, err := client.SynthesizeSpeech(context.Background(), &req)
	if err != nil {
		return
	}
	opusFile, err = NewOpusFile(bytes.NewBuffer(resp.AudioContent))
	return
}
