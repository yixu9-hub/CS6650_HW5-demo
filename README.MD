## Instructions

### Prepare Credentials

Retrieve you temporary credentials from Learner's Lab.
Enter your configuration when prompted:
```
aws configure
```

And set your session token:
```
aws configure set aws_session_token <YOUR-TEMP-SESSION-TOKEBN>
```

### Apply Infrastructure
```
cd terraform
terraform init
terraform apply -auto-approve
```

### Send Requests
Make sure you are in terraform folder

Get public ip address
```
aws ec2 describe-network-interfaces \
--network-interface-ids $(
    aws ecs describe-tasks \
    --cluster $(terraform output -raw ecs_cluster_name) \
    --tasks $(
        aws ecs list-tasks \
        --cluster $(terraform output -raw ecs_cluster_name) \
        --service-name $(terraform output -raw ecs_service_name) \
        --query 'taskArns[0]' --output text
    ) \
    --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" \
    --output text
) \
--query 'NetworkInterfaces[0].Association.PublicIp' \
--output text
```
powershell:
$cluster = terraform output -raw ecs_cluster_name
$service = terraform output -raw ecs_service_name
$taskArn = aws ecs list-tasks --cluster $cluster --service-name $service --query 'taskArns[0]' --output text
$eni = aws ecs describe-tasks --cluster $cluster --tasks $taskArn --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" --output text
aws ec2 describe-network-interfaces --network-interface-ids $eni --query 'NetworkInterfaces[0].Association.PublicIp' --output text

Send some requests:
```
curl http://<PUBLIC-IP-ADDRESS>:8080/healthz
curl -sS -X GET "http://<PUBLIC-IP>:8080/products" -H "Accept: application/json"
curl -sS -X POST "http://<PUBLIC-IP>:8080/products" \
  -H "Content-Type: application/json" \
  -d '{"id":"1","name":"Example Product","description":"demo","price":9.99}'
```

## Clean Up
```
terraform destroy -auto-approve
```