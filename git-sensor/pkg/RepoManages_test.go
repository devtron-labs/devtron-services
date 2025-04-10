package pkg

import (
	"github.com/devtron-labs/git-sensor/pkg/git"
	"reflect"
	"testing"
)

func TestRepoManagerImpl_GetWebhookAndCiDataById(t *testing.T) {
	type args struct {
		id                   int
		ciPipelineMaterialId int
	}
	tests := []struct {
		name    string
		args    args
		want    *git.WebhookAndCiData
		wantErr bool
	}{
		
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl :=
			got, err := impl.GetWebhookAndCiDataById(tt.args.id, tt.args.ciPipelineMaterialId)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWebhookAndCiDataById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWebhookAndCiDataById() got = %v, want %v", got, tt.want)
			}
		})
	}
}
