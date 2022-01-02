package main

import (
  "context"
  "flag"
  "fmt"
  "log"
  "main/aws"
  "time"
)

func initRunConfig() *aws.RunConfig {
  groupName := flag.String("group", "", "the name of the Auto Scaling group to update; required")
  imageID := flag.String("image", "", "AMI id to update the group; optional: do not use if you want to create an AMI from a running instance.")
  instanceID := flag.String("instance", "", "AWS EC2 instance ID to create the AMI from; optional: do not use if you already have an AMI ID.")
  amiName := flag.String("ami", "", "the name of the AMI to be created from the selected instance; optional: use together with the `--instance` argument only.")
  updateTimeoutStr := flag.String("update-timeout", "30m", "the time limit to complete the instance refresh; optional: the default is 30 minutes. Use the Golang duration strings to override, see https://pkg.go.dev/time#ParseDuration.")
  updateTickStr := flag.String("update-tick", "1m", "the time between status updates in the log file; optional: the default is one minute. Making this parameter lower might speed up the overall execution. Use the Golang duration strings to override, see https://pkg.go.dev/time#ParseDuration.")
  flag.Parse()

  updateTimeout, err := time.ParseDuration(*updateTimeoutStr)
  if err != nil {
    log.Fatalf("cannot parse the update timeout string: %v", err)
  }
  updateTick, err := time.ParseDuration(*updateTickStr)
  if err != nil {
    log.Fatalf("cannot parse the update tick string: %v", err)
  }

  return &aws.RunConfig{
    AMIName:       *amiName,
    ImageID:       *imageID,
    InstanceID:    *instanceID,
    GroupName:     *groupName,
    UpdateTimeout: updateTimeout,
    UpdateTick:    updateTick,
  }
}

func validate(rc *aws.RunConfig) error {
  if rc.GroupName == "" {
    return fmt.Errorf("the group name should be specified")
  }
  if rc.InstanceID != "" && rc.AMIName == "" {
    return fmt.Errorf("AMI name should be specified when creatint an AMI from instance")
  }
  if rc.InstanceID == "" && rc.ImageID == "" {
    return fmt.Errorf("either instance ID or image ID should be specified")
  }
  if rc.InstanceID != "" && rc.ImageID != "" {
    return fmt.Errorf("instance ID and image ID shouldn't be specified simultaneously")
  }
  return nil
}

func main() {
  rc := initRunConfig()
  if err := validate(rc); err != nil {
    log.Fatalln(err)
  }
  client, err := aws.NewClient(context.Background(), rc)
  if err != nil {
    log.Fatalf("cannot initialize the AWS client: %v", err)
  }
  if rc.InstanceID != "" {
    log.Printf("creating image %q from the instance %s", rc.AMIName, rc.InstanceID)
    rc.ImageID, err = client.CreateAMI(rc.InstanceID, rc.AMIName)
    if err != nil {
      log.Fatalf("cannot create AMI from the instance: %v", err)
    }
    log.Printf("created image %q from instance %s", rc.ImageID, rc.InstanceID)
  }
  if err := client.UpdateLaunchTemplate(); err != nil {
    log.Fatalf("cannot update the launch template: %v", err)
  }
  log.Printf("will run instance refresh for the auto scale group %q", rc.GroupName)
  if err := client.RunInstanceRefresh(); err != nil {
    log.Fatalf("cannot execute instance refresh: %v", err)
  }
  log.Printf("instance refresh for the auto scale group %q completed successfully", rc.GroupName)
}
