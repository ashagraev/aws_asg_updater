package aws

import (
  "fmt"
  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/service/autoscaling"
  "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
  "log"
  "time"
)

type refreshStatus struct {
  CompletionPercent int32
  Status            types.InstanceRefreshStatus
}

func (c *Client) getInstanceRefreshStatus(refreshID string) (*refreshStatus, error) {
  refreshDescriptions, err := c.autoscalingClient.DescribeInstanceRefreshes(c.ctx, &autoscaling.DescribeInstanceRefreshesInput{
    AutoScalingGroupName: aws.String(c.rc.GroupName),
    InstanceRefreshIds:   []string{refreshID},
  })
  if err != nil {
    return nil, fmt.Errorf("cannot get refresh status for autoscaling group %s: %v", refreshID, err)
  }
  if len(refreshDescriptions.InstanceRefreshes) != 1 {
    return nil, fmt.Errorf("got %d != 1 refreshes for group %s with refresh id %s", len(refreshDescriptions.InstanceRefreshes), c.rc.GroupName, refreshID)
  }
  return &refreshStatus{
    CompletionPercent: *refreshDescriptions.InstanceRefreshes[0].PercentageComplete,
    Status:            refreshDescriptions.InstanceRefreshes[0].Status,
  }, nil
}

func (c *Client) awaitInstanceRefreshCompletion(refreshID string, targetStatus types.InstanceRefreshStatus) (*refreshStatus, error) {
  finishTime := time.Now().Add(c.rc.UpdateTimeout)
  for time.Now().Before(finishTime) {
    time.Sleep(c.rc.UpdateTick)
    status, err := c.getInstanceRefreshStatus(refreshID)
    if err != nil {
      log.Printf("cannot get instance refresh status: %v", err)
      continue
    }
    log.Printf("updating group %s, update id %s: %s, %d%% completion", c.rc.GroupName, refreshID, status.Status, status.CompletionPercent)
    if status.Status != types.InstanceRefreshStatusPending && status.Status != types.InstanceRefreshStatusInProgress {
      if status.Status == targetStatus {
        return status, nil
      }
      return nil, fmt.Errorf("updating group %s, update id %s: wrong finish status %s, expected %s", c.rc.GroupName, refreshID, status.Status, targetStatus)
    }
  }
  return nil, fmt.Errorf("instance refresh deadline %v reached", c.rc.UpdateTimeout)
}

func (c *Client) RunInstanceRefresh() error {
  refreshOutput, err := c.autoscalingClient.StartInstanceRefresh(c.ctx, &autoscaling.StartInstanceRefreshInput{
    AutoScalingGroupName: aws.String(c.rc.GroupName),
  })
  if err != nil {
    return fmt.Errorf("cannot start instance refresh for autoscaling group %s: %v", c.rc.GroupName, err)
  }
  _, err = c.awaitInstanceRefreshCompletion(*refreshOutput.InstanceRefreshId, types.InstanceRefreshStatusSuccessful)
  if err != nil {
    log.Printf("instance refresh %s for autoscaling group %s not successful, cancelling: %v", *refreshOutput.InstanceRefreshId, c.rc.GroupName, err)
    cancelOutput, err := c.autoscalingClient.CancelInstanceRefresh(c.ctx, &autoscaling.CancelInstanceRefreshInput{
      AutoScalingGroupName: aws.String(c.rc.GroupName),
    })
    if err != nil {
      return fmt.Errorf("cannot cancel instance refresh for autoscaling group %s: %v", c.rc.GroupName, err)
    }
    _, err = c.awaitInstanceRefreshCompletion(*cancelOutput.InstanceRefreshId, types.InstanceRefreshStatusCancelled)
    if err != nil {
      return fmt.Errorf("cannot cancel instance refresh %s for autoscaling group %s: %v", *cancelOutput.InstanceRefreshId, c.rc.GroupName, err)
    }
    log.Printf("cancelled instance refresh %s for autoscaling group %s", *refreshOutput.InstanceRefreshId, c.rc.GroupName)
    return fmt.Errorf("instance refresh %s for autoscaling group %s not successful: %v", *refreshOutput.InstanceRefreshId, c.rc.GroupName, err)
  }
  return nil
}
