package alerts

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/spf13/cobra"
)

type ID struct {
	ID string `json:"id"`
}

type Matchers struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Silence struct {
	ID      string     `json:"id"`
	Matchers []Matchers `json:"matchers"`
}

func NewCmdListSilence() *cobra.Command {
	return &cobra.Command{
		Use:               "list-silence <cluster-id>",
		Short:             "List all silences for given alert",
		Long:              `list all  silence`,
		Args:              cobra.ExactArgs(1),
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			ListSilence(args[0])
		},
	}
}

// osdctl alerts list-silence ${CLUSTERID}
func ListSilence(clusterID string){
	var silence []Silence

	silenceCmd := []string{"amtool", "silence", "--alertmanager.url", LocalHostUrl, "-o", "json"}

	kubeconfig, clientset, err := GetKubeConfigClient(clusterID)
	if err != nil {
		log.Fatal(err)
	}

	op, err := ExecInPod(kubeconfig, clientset, silenceCmd)
	if err != nil {
		fmt.Println(err)
	}

	opSlice := []byte(op)
	err = json.Unmarshal(opSlice, &silence)
	if err != nil {
		fmt.Println("Error in unmarshaling the data", err)
	}

	for _, s := range silence {
		id, matchers := s.ID, s.Matchers
		for _, matcher := range matchers{
			fmt.Printf("Found %v %v with silence id %s\n", matcher.Name, matcher.Value, id) 
		}
	}
	fmt.Println("No silences found, all silence has been cleared.")
}
