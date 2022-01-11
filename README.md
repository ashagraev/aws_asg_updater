# AWS Auto Scaling Groups Updater

[AWS Auto Scaling group](https://docs.aws.amazon.com/autoscaling/ec2/userguide/AutoScalingGroup.html) is a great way of
managing Amazon EC2 instances. AWS Auto Scaling group watches the corresponding instances' health and launches new
instances whenever needed: to replace instances that became unhealthy or to [scale the group](
https://docs.aws.amazon.com/autoscaling/ec2/userguide/scaling_plan.html). Furthermore, with an [AWS Elastic Load
Balancer](https://docs.aws.amazon.com/autoscaling/ec2/userguide/attach-load-balancer-asg.html) attached, AWS Auto
Scaling Group also takes ELB health checks into account, making it possible to replace the instances based on
service-level signals.

However, when it comes to updating services, AWS Auto Scaling Groups become tricky. The recommended way is to create
the group using a [launch template](https://docs.aws.amazon.com/autoscaling/ec2/userguide/create-asg-launch-template.html)
containing the information required to launch new instances, such as instance type and [Amazon Machine Image](
https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AMIs.html) (AMI). Hence, to update the service, one needs to create
a new AMI, update the corresponding launch template, and execute an [instance refresh](
https://docs.aws.amazon.com/autoscaling/ec2/userguide/asg-instance-refresh.html).

This is where AWS Auto Scaling Groups Updater comes in handy. It uses [AWS EC2 API](
https://docs.aws.amazon.com/AWSEC2/latest/APIReference/Welcome.html) and [AWS EC2 Auto Scaling API](
https://docs.aws.amazon.com/autoscaling/ec2/APIReference/Welcome.html) to perform the following operations
automatically:
- register a new AMI from a running instance;
- create a new version of the Auto Scaling group's launch template with this new AMI attached;
- execute the instance refresh for the Auto Scaling group;
- cancel the instance refresh if it's unable to finish within the desired time;
- handle all the errors that might happen along the way.

The tool uses the default AWS credentials config. Run `aws configure` or set up the environment variables
`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION` before running the tool.

## Examples of Use

The program options to create a new AMI from the instance `i-0d8767eed40de1728`, name it `service_ami`,
update the Auto Scaling group `my_service_group` with it:

`aws_asg_updater --group my_service_group --instance i-0d8767eed40de1728 --ami service_ami`

The output looks like that then:

```
2022/01/02 00:23:02 creating image "service_ami" from the instance i-0d8767eed40de1728
2022/01/02 00:24:03 creating image service_ami (ami-070a1f039b2945569): pending
2022/01/02 00:25:04 creating image service_ami (ami-070a1f039b2945569): pending
2022/01/02 00:26:04 creating image service_ami (ami-070a1f039b2945569): available
2022/01/02 00:26:04 created image "ami-070a1f039b2945569" from instance i-0d8767eed40de1728
2022/01/02 00:26:05 launch template id: lt-0b049e246796cc5cd
2022/01/02 00:26:05 latest launch template version id: 19
2022/01/02 00:26:06 created new version 20 for the launch template lt-0b049e246796cc5cd
2022/01/02 00:26:06 set the version 20 for the launch template lt-0b049e246796cc5cd as default
2022/01/02 00:26:06 will run instance refresh for the auto scale group "my_service_group"
2022/01/02 00:27:07 updating group my_service_group, update id 1608d67f-931c-4e83-8aff-56721c95e86a: InProgress, 0% completion
2022/01/02 00:28:08 updating group my_service_group, update id 1608d67f-931c-4e83-8aff-56721c95e86a: InProgress, 0% completion
2022/01/02 00:29:09 updating group my_service_group, update id 1608d67f-931c-4e83-8aff-56721c95e86a: InProgress, 0% completion
2022/01/02 00:30:10 updating group my_service_group, update id 1608d67f-931c-4e83-8aff-56721c95e86a: InProgress, 0% completion
2022/01/02 00:31:11 updating group my_service_group, update id 1608d67f-931c-4e83-8aff-56721c95e86a: Successful, 100% completion
2022/01/02 00:31:11 instance refresh for the auto scale group "my_service_group" completed successfully
```

If you already have an AMI ID, just skip the `--ami` and `--instance` options, and use the `--image` option, e.g.:

`aws_asg_updater --group my_service_group --image ami-070a1f039b2945569`

This will update the launch template and run the instance refresh:

```
2022/01/02 11:36:09 launch template id: lt-0b049e246796cc5cd
2022/01/02 11:36:10 latest launch template version id: 20
2022/01/02 11:36:10 created new version 21 for the launch template lt-0b049e246796cc5cd
2022/01/02 11:36:11 set the version 21 for the launch template lt-0b049e246796cc5cd as default
2022/01/02 11:36:11 will run instance refresh for the auto scale group "appscience_bot"
2022/01/02 11:37:12 updating group appscience_bot, update id 197f64f8-9b1e-4a9c-b4ad-85b7882ca170: InProgress, 0% completion
2022/01/02 11:38:13 updating group appscience_bot, update id 197f64f8-9b1e-4a9c-b4ad-85b7882ca170: InProgress, 0% completion
2022/01/02 11:39:14 updating group appscience_bot, update id 197f64f8-9b1e-4a9c-b4ad-85b7882ca170: InProgress, 0% completion
2022/01/02 11:40:15 updating group appscience_bot, update id 197f64f8-9b1e-4a9c-b4ad-85b7882ca170: InProgress, 0% completion
2022/01/02 11:41:15 updating group appscience_bot, update id 197f64f8-9b1e-4a9c-b4ad-85b7882ca170: InProgress, 0% completion
2022/01/02 11:42:16 updating group appscience_bot, update id 197f64f8-9b1e-4a9c-b4ad-85b7882ca170: Successful, 100% completion
2022/01/02 11:42:16 instance refresh for the auto scale group "appscience_bot" completed successfully
```

## Program Arguments

- `group`: the name of the Auto Scaling group to update; required.
- `image`: AMI id to update the group; optional: do not use if you want to create an AMI from a running instance.
- `instance`: AWS EC2 instance ID to create the AMI from; optional: do not use if you already have an AMI ID.
- `ami`: the name of the AMI to be created from the selected instance; optional: use together with the `--instance` argument only.
- `update-timeout`: the time limit to complete the instance refresh; optional: the default is 30 minutes. Use the Golang duration strings to override, see https://pkg.go.dev/time#ParseDuration. 
- `update-tick`: the time between status updates in the log file; optional: the default is one minute. Making this parameter lower might speed up the overall execution. Use the Golang duration strings to override, see https://pkg.go.dev/time#ParseDuration.

## Installation

Assuming you already have Golang installed on the machine, simply run:

```
go install github.com/ashagraev/aws_asg_updater@latest
```

The tool will then appear in your golang binary folder (e.g., `~/go/bin/aws_asg_updater`). If you don't have Golang yet,
consider installing it using the official guide https://go.dev/doc/install.

Alternatively, you can download the pre-built binaries from the latest release:
https://github.com/ashagraev/aws_asg_updater/releases/latest.
