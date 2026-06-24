const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Load .env
const envPath = path.join(__dirname, '..', '.env');
if (fs.existsSync(envPath)) {
  const envConfig = fs.readFileSync(envPath, 'utf8');
  envConfig.split('\n').forEach(line => {
    const match = line.match(/^([^=]+)=(.*)$/);
    if (match) {
      let val = match[2].trim();
      if (val.startsWith('"') && val.endsWith('"')) val = val.slice(1, -1);
      process.env[match[1].trim()] = val;
    }
  });
}

const jules = require('./jules_client');
const tasks = require('./new_tasks');

const MAX_CONCURRENCY = 40;

const PERSONAS = [
  'apollo.md',
  'athena.md',
  'hermes.md',
  'artemis.md',
  'hephaestus.md'
];

const VISION_BLOCK = `
GLOBAL VISION:
200% Accuracy Required. Ensure full GDPR compliance, strict typing, extensive regression testing, and Product Hunt-level UI/UX polish.

CRITICAL REFERENCES:
Please reference and use the following repositories for architectural patterns, core behaviors, and system design while executing this task:
1. https://github.com/open-gsd/gsd-core
2. https://github.com/snarktank/ralph
`;

function runMerge() {
  console.log('Running merge_all.js...');
  try {
    const mergeScriptPath = path.join(__dirname, 'merge_all.js');
    execSync(`node "${mergeScriptPath}"`, { stdio: 'inherit' });
  } catch (e) {
    console.error('Merge script failed', e.message);
  }
}

async function orchestrate() {
  runMerge();

  let sessions = [];
  try {
    const res = await jules.getSessions();
    sessions = res.sessions || [];
    console.log(`Fetched ${sessions.length} total sessions from Jules.`);
  } catch (e) {
    console.error('Failed to fetch sessions:', e.message);
    return;
  }

  let activeCount = 0;

  for (const session of sessions) {
    if (session.status === 'COMPLETED' || session.status === 'FAILED') {
      continue;
    }
    activeCount++;
  }
  console.log(`Currently ${activeCount} active sessions.`);
  activeCount = 0; // reset to count again in the real loop

  for (const session of sessions) {
    if (session.status === 'COMPLETED' || session.status === 'FAILED') {
      continue;
    }
    activeCount++;

    if (session.status === 'AWAITING_PLAN_APPROVAL') {
      console.log(`Approving plan for session ${session.id}`);
      await jules.approvePlan(session.id);
      await jules.sendMessage(session.id, "The plan looks solid. Please proceed.");
    } 
    else if (session.status === 'FAILED') {
      console.log(`Session ${session.id} failed. Investigating...`);
      const activities = await jules.getSessionActivities(session.id);
      const reason = (activities && activities.error) ? activities.error : "Unknown reason";
      await jules.sendMessage(session.id, `I see the execution failed with reason: ${reason}. Please analyze the logs, fix the root cause, and continue.`);
    }
    else if (session.status === 'AWAITING_USER_FEEDBACK') {
      console.log(`Sending unblock message to session ${session.id}`);
      await jules.sendMessage(session.id, "Please proceed autonomously using your best judgment. Do not block on me.");
    }
  }

  const availableSlots = MAX_CONCURRENCY - activeCount;
  if (availableSlots > 0 && tasks.length > 0) {
    console.log(`Spawning ${Math.min(availableSlots, tasks.length)} new tasks...`);
    
    for (let i = 0; i < availableSlots; i++) {
      if (tasks.length === 0) break;
      const task = tasks[0]; // Peek at task

      const randomPersonaFile = PERSONAS[Math.floor(Math.random() * PERSONAS.length)];
      const personaPath = path.join(__dirname, '..', '.jules', randomPersonaFile);
      let personaContent = '';
      if (fs.existsSync(personaPath)) {
        personaContent = fs.readFileSync(personaPath, 'utf8');
      }

      const prompt = `${task}\n\n${personaContent}\n\n${VISION_BLOCK}`;
      const title = `Implement task: ${task.substring(0, 30)}...`;

      console.log(`Spawning session for: ${title}`);
      try {
        await jules.createSession(prompt, 'github/goyalmunish85/ScrumMaster.ai', 'main', title);
        tasks.shift(); // Only remove if successful
      } catch (e) {
        console.error('Failed to spawn session:', e.message);
        break; // Stop trying to spawn if we hit an API error (like rate limit)
      }
    }

    // Persist updated tasks back to new_tasks.js so we don't repeat them
    const updatedTasksContent = `module.exports = ${JSON.stringify(tasks, null, 2)};\n`;
    fs.writeFileSync(path.join(__dirname, 'new_tasks.js'), updatedTasksContent, 'utf8');
  }
}

orchestrate().then(() => console.log('Orchestrator run complete.')).catch(console.error);
