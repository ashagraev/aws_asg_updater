package aws

import (
  "fmt"
  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/service/autoscaling"
  "github.com/aws/aws-sdk-go-v2/service/ec2"
  "github.com/aws/aws-sdk-go-v2/service/ec2/types"
  "log"
)

func (c *Client) getLaunchTemplateID() (string, error) {
  g, err := c.autoscalingClient.DescribeAutoScalingGroups(c.ctx, &autoscaling.DescribeAutoScalingGroupsInput{
    AutoScalingGroupNames: []string{c.rc.GroupName},
  })
  if err != nil {
    return "", fmt.Errorf("cannot get description of the autoscaling group %s: %v", c.rc.GroupName, err)
  }
  if len(g.AutoScalingGroups) != 1 {
    return "", fmt.Errorf("got wrong number %d != 1 of autoscaling groups for id %s", len(g.AutoScalingGroups), c.rc.GroupName)
  }
  if g.AutoScalingGroups[0].LaunchTemplate == nil {
    return "", fmt.Errorf("group %s has no launch templates", c.rc.GroupName)
  }
  return *g.AutoScalingGroups[0].LaunchTemplate.LaunchTemplateId, nil
}

func (c *Client) UpdateLaunchTemplate() error {
  launchTemplateID, err := c.getLaunchTemplateID()
  log.Printf("launch template id: %s", launchTemplateID)
  if err != nil {
    return err
  }
  launchTemplates, err := c.ec2Client.DescribeLaunchTemplates(c.ctx, &ec2.DescribeLaunchTemplatesInput{
    LaunchTemplateIds: []string{launchTemplateID},
  })
  if err != nil {
    return fmt.Errorf("cannot get data for the launch template %s: %v", launchTemplateID, err)
  }
  if len(launchTemplates.LaunchTemplates) != 1 {
    return fmt.Errorf("got wrong number %d != 1 of launch templates for id %s", len(launchTemplates.LaunchTemplates), c.rc.GroupName)
  }
  latestVersion := *launchTemplates.LaunchTemplates[0].LatestVersionNumber
  log.Printf("latest launch template version id: %d", latestVersion)
  createLaunchTemplateOutput, err := c.ec2Client.CreateLaunchTemplateVersion(c.ctx, &ec2.CreateLaunchTemplateVersionInput{
    LaunchTemplateData: &types.RequestLaunchTemplateData{
      ImageId: aws.String(c.rc.ImageID),
    },
    LaunchTemplateId: aws.String(launchTemplateID),
    SourceVersion:    aws.String(fmt.Sprintf("%d", latestVersion)),
  })
  if err != nil {
    return fmt.Errorf("cannot create launch template version for launch template %s: %v", launchTemplateID, err)
  }
  log.Printf("created new version %d for the launch template %s", *createLaunchTemplateOutput.LaunchTemplateVersion.VersionNumber, launchTemplateID)
  _, err = c.ec2Client.ModifyLaunchTemplate(c.ctx, &ec2.ModifyLaunchTemplateInput{
    DefaultVersion:   aws.String(fmt.Sprintf("%v", *createLaunchTemplateOutput.LaunchTemplateVersion.VersionNumber)),
    LaunchTemplateId: aws.String(launchTemplateID),
  })
  if err != nil {
    return fmt.Errorf("cannot modify the default version for launch template %s: %v", launchTemplateID, err)
  }
  log.Printf("set the version %d for the launch template %s as default", *createLaunchTemplateOutput.LaunchTemplateVersion.VersionNumber, launchTemplateID)
  return nil
}
