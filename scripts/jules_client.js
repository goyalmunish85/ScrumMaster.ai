const https = require('https');

const API_KEY = process.env.JULES_API_KEY;
const BASE_URL = 'jules.googleapis.com'; // Official Jules API

function request(method, path, data = null) {
  return new Promise((resolve, reject) => {
    if (!API_KEY) {
      return reject(new Error('JULES_API_KEY is not set in environment'));
    }

    const options = {
      hostname: BASE_URL,
      path: path,
      method: method,
      headers: {
        'x-goog-api-key': API_KEY,
        'Content-Type': 'application/json'
      }
    };

    const req = https.request(options, (res) => {
      let responseBody = '';
      res.on('data', chunk => { responseBody += chunk; });
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          try {
            resolve(JSON.parse(responseBody));
          } catch(e) {
            resolve(responseBody);
          }
        } else {
          reject(new Error(`API Error: ${res.statusCode} - ${responseBody}`));
        }
      });
    });

    req.on('error', (e) => reject(e));

    if (data) {
      req.write(JSON.stringify(data));
    }
    req.end();
  });
}

async function createSession(prompt, repo, branch, title) {
  const payload = {
    prompt: prompt,
    sourceContext: {
      source: `sources/${repo}`, // The API expects something like "sources/github/owner/repo"
      githubRepoContext: {
        startingBranch: branch
      }
    }
  };
  return request('POST', '/v1alpha/sessions', payload);
}

async function getSessions() {
  return request('GET', '/v1alpha/sessions');
}

async function getSessionActivities(sessionId) {
  return request('GET', `/v1alpha/sessions/${sessionId}/activities`);
}

async function approvePlan(sessionId) {
  return request('POST', `/v1alpha/sessions/${sessionId}/approve`);
}

async function sendMessage(sessionId, message) {
  return request('POST', `/v1alpha/sessions/${sessionId}/messages`, { message });
}

module.exports = {
  createSession,
  getSessions,
  getSessionActivities,
  approvePlan,
  sendMessage
};
