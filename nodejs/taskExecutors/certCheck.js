'use strict'
const https = require('https')

module.exports.handle = async taskConfig => {
  return await checkCert(taskConfig.host)
}

async function checkCert(host) {
  return new Promise((resolve, reject) => {
    const options = {
      host: host,
      method: 'GET',
      rejectUnauthorized: false,
    };

    const req = https.request(options, function(res) {
      const now = new Date();
      const millisecondsPerDay = 24 * 60 * 60 * 1000
      const certificateInfo = res.connection.getPeerCertificate();

      const validFrom = new Date(certificateInfo.valid_from);
      const validTo = new Date(certificateInfo.valid_to);
      const daysLeft = Math.floor((validTo - now) / millisecondsPerDay)
      const expired = daysLeft <= 0;

      const result = {
        validFrom: validFrom.toISOString(),
        validTo: validTo.toISOString(),
        daysLeft: daysLeft,
        expired: expired,
        host: host
      };

      resolve(result)
    });

    req.on('error', (err) => {
      reject(err)
    });

    req.end();
  })
}
