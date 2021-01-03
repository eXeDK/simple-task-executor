'use strict'
const ssllabs = require('node-ssllabs')

module.exports.handle = async taskConfig => {
  const options = {
    host: taskConfig.host,
    maxAge: 2
  }

  return await scanHost(options)
}

async function scanHost(options) {
  return new Promise((resolve, reject) => {
    const results = {}

    ssllabs.scan(options, (err, host) => {
      if (err !== null) {
        reject(err)
        return
      }

      results.engineVersion = host.engineVersion
      results.criteriaVersion = host.criteriaVersion
      results.endpoints = {}

      host.endpoints.forEach(function (endpoint) {
        results.endpoints[endpoint.ipAddress] = {
          grade: endpoint.grade,
          ipAddress: endpoint.ipAddress,
          serverName: endpoint.serverName,
          gradeTrustIgnored: endpoint.gradeTrustIgnored,
          duration: endpoint.duration
        }
      })

      resolve(results)
    })
  })
}
