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
  amiName := flag.String("ami", "", "AMI name to create from the instance")
  imageID := flag.String("image", "", "image id to be used in the autoscaling group")
  instanceID := flag.String("instance", "", "source instance id to create the image from")
  groupName := flag.String("group", "", "auto scale group name")
  updateTimeoutStr := flag.String("update-timeout", "30m", "update timeout (e.g., 30m, 1h, 7h20m7s)")
  updateTickStr := flag.String("update-tick", "1m", "update tick (e.g., 1m, 5m, 10s, 7h20m7s)")
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
