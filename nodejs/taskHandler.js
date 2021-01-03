'use strict'
const AWS = require('aws-sdk')
const { v4: uuidv4 } = require('uuid')

// Import task executors
const certCheck = require('taskExecutors/certCheck')
const httpPing = require('taskExecutors/httpPing')
const sslLabsCheck = require('taskExecutors/sslLabsCheck')

// Needed AWS services
const dynamodb = new AWS.DynamoDB.DocumentClient()
const sqs = new AWS.SQS()

// Environment variables
const queueUrl = process.env.QUEUE_URL

module.exports.main = async event => {
  // Get targetId and config
  const targetId = getTargetId(event)
  const isTestInvoke = getTestInvoke(event)
  const targetConfig = await getTargetConfig(targetId)

  // If we are not active, stop the task
  if (targetConfig.active === false && isTestInvoke === false) {
    console.info('Task stopped - target not active')
    return
  }

  // Execute logic for target
  const taskResult = await executeTask(targetConfig)

  // Save task response
  await saveTaskResult(targetId, taskResult)

  // Schedule a new ping task
  if (isTestInvoke === false) {
    await scheduleTask(queueUrl, targetId, targetConfig)
  }
}

function getTargetId(event) {
  return event.Records[0].messageAttributes.TargetId.stringValue
}

function getTestInvoke(event) {
  return Object.keys(event.Records[0].messageAttributes).includes('TestInvoke') && event.Records[0].messageAttributes.TestInvoke.stringValue === "true"
}

async function getTargetConfig(targetId) {
  // Get config from dynamodb
  const params = {
    TableName: 'Targets',
    Key: {
      id: targetId
    },
    ProjectionExpression: 'active, config, delay, targetType'
  }

  return await dynamodb.get(params).promise()
      .then(data => data.Item)
}

async function executeTask(targetConfig) {
  const taskConfig = JSON.parse(targetConfig.config)

  try {
    switch (targetConfig.targetType) {
      case 'certCheck':
        return await certCheck.handle(taskConfig)

      case 'httpPing':
        return await httpPing.handle(taskConfig)

      case 'sslLabsCheck':
        return await sslLabsCheck.handle(taskConfig)

      default:
        return null
    }
  } catch (err) {
    console.error('Task executor sent a rejection. Returning null', err)
    return null
  }
}

async function saveTaskResult(targetId, taskResult) {
  // If result is null, then do not save
  if (taskResult === null) {
    console.info('saveTaskResult result is null for target "' + targetId + '"')
    return
  }

  // Save task result to dynamodb
  const params = {
    TableName: 'TargetTaskResults',
    Item: {
      id: uuidv4(),
      timestamp: Date.now(),
      targetId: targetId,
      result: taskResult
    }
  }

  return dynamodb.put(params).promise()
}

async function scheduleTask(targetId, targetConfig) {
  // Extract relevant info from event
  const delay = targetConfig.delay

  // Schedule a new ping task
  const sendMessageRequest = {
    QueueUrl: queueUrl,
    MessageBody: '{}',
    DelaySeconds: delay,
    MessageAttributes: {
      TargetId: {
        DataType: 'String',
        StringValue: targetId
      }
    }
  }
  console.info('New task scheduled for target "' + targetId + '"')

  return await sqs.sendMessage(sendMessageRequest).promise()
}