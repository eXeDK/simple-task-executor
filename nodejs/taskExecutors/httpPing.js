'use strict'
const got = require('got')

module.exports.handle = async taskConfig => {
  try {
    const response = await got(taskConfig.url, {
      http2: true,
      throwHttpErrors: false,
      cache: false,
      followRedirect: false,
      decompress: false
    })

    return {
      timing: response.timings
    }
  } catch (error) {
    return null
  }
}
