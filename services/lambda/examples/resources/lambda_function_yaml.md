**YAML**

```yaml
resources:
  processOrdersFunction:
	type: aws/lambda/function
	metadata:
	  displayName: Order Processing Function
	  description: This function processes customer orders.
	  labels:
	    app: orders
	spec:
	  functionName: orders-ProcessOrdersFunction-v1
      # The zip file containing the Lambda function code in this case is stored in an S3 bucket,
      # the code can also be provided inline using the `zipFile` property.
	  code:
	    s3Bucket: my-bucket
	    s3Key: order-processing.zip
	  role: arn:aws:iam::123456789012:role/lambda-execution-role
	  handler: index.handler
	  runtime: nodejs22.x
	  memorySize: 256
	  timeout: 30
	  environment:
	    variables:
	      ORDER_QUEUE_URL: https://sqs.us-east-1.amazonaws.com/123456789012/order-queue
	  tracingConfig:
	    mode: Active
	  vpcConfig:
	    securityGroupIds:
	      - sg-0123456789abcdef0
	      - sg-0fedcba9876543210
	    subnetIds:
	      - subnet-0123456789abcdef0
	      - subnet-0fedcba9876543210
	  tags:
	    Environment: Production
	    Application: OrderProcessing
```
