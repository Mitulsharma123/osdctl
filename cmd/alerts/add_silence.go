package alerts

import (
	"fmt"
	"log"
	"strings"

	ocmutils "github.com/openshift/osdctl/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type addSilenceCmd struct {
	clusterID string
	alertID   []string
	duration  string
	comment   string
	all       bool
}

func NewCmdAddSilence() *cobra.Command {
	addSilenceCmd := &addSilenceCmd{}
	cmd := &cobra.Command{
		Use:               "add-silence <cluster-id> [--all --duration --comment | --alertname --duration --comment]",
		Short:             "Add a new silence for alert present in the cluster",
		Long:              `add new silence for specfic or all alert with comment and duration of alert`,
		Args:              cobra.ExactArgs(1),
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			addSilenceCmd.clusterID = args[0]
			AddSilence(addSilenceCmd)
		},
	}

	cmd.Flags().StringSliceVar(&addSilenceCmd.alertID, "alertname", []string{}, "alertname (comma-separated)")
	cmd.Flags().StringVarP(&addSilenceCmd.comment, "comment", "c", "", "add comment about silence")
	cmd.Flags().StringVarP(&addSilenceCmd.duration, "duration", "d", "15d", "add duration for silence")
	//By default duration should be 15 days
	cmd.Flags().BoolVarP(&addSilenceCmd.all, "all", "a", false, "add silences for all alert")
	//--all Flag to add silences

	return cmd
}

func AddSilence(cmd *addSilenceCmd) {
	username := GetUserInfo()
	clusterID := cmd.clusterID
	alertID := cmd.alertID
	comment := cmd.comment
	duration := cmd.duration
	all := cmd.all

	kubeconfig, clientset, err := GetKubeConfigClient(clusterID)
	if err != nil {
		log.Fatal(err)
	}

	if all {
		AddAllSilence(clusterID, duration, comment, username, kubeconfig, clientset)
	} else if len(alertID) > 0 {
		AddAlertNameSilence(alertID, duration, comment, username, kubeconfig, clientset)
	} else {
		fmt.Println("No valid option specified. Use --all or --alertname.")
	}

}

func AddAllSilence(clusterID, duration, comment, username string, kubeconfig *rest.Config, clientset *kubernetes.Clientset) {
	addCmd := []string{
		"amtool",
		"silence",
		"add",
		"cluster=" + clusterID,
		"--alertmanager.url=" + LocalHostUrl,
		"--duration=" + duration,
		"--comment=" + comment,
	}

	output, err := ExecInPod(kubeconfig, clientset, addCmd)
	if err != nil {
		fmt.Println(err)
	}

	formattedOutput := strings.Replace(output, "\n", " ", -1)

	fmt.Printf("All alerts for cluster %s has been silenced with id \"%s\"for a duration of %s by user %s.\n", ClusterName, formattedOutput, duration, username)
}

func AddAlertNameSilence(alertID []string, duration, comment, username string, kubeconfig *rest.Config, clientset *kubernetes.Clientset) {
	for _, alertname := range alertID {
		addCmd := []string{
			"amtool",
			"silence",
			"add",
			"alertname=" + alertname,
			"--alertmanager.url=" + LocalHostUrl,
			"--duration=" + duration,
			"--comment=" + comment,
		}

		output, err := ExecInPod(kubeconfig, clientset, addCmd)
		if err != nil {
			fmt.Println(err)
		}

		formattedOutput := strings.Replace(output, "\n", " ", -1)

		fmt.Printf("Alert \"%s\" has been silenced with id \"%s\"for duration of %s by user %s.\n", alertname, formattedOutput, duration, username)
	}
}

// User name of who added the silence
func GetUserInfo() string {
	connection, err := ocmutils.CreateConnection()
	if err != nil {
		fmt.Printf("Error %s in create connection.", err)
	}
	defer connection.Close()

	account, err := connection.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		fmt.Printf("Error %s in retreving the details of account.", err)
	}

	name, _ := account.Body().GetUsername()
	return name
}
