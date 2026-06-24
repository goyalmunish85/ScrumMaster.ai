const { execSync } = require('child_process');

function run(command) {
  try {
    return execSync(command, { encoding: 'utf8', stdio: 'pipe' });
  } catch (e) {
    return null;
  }
}

function mergeAll() {
  console.log('Fetching all branches...');
  run('git fetch --all');

  const branchesOutput = run('git branch -r');
  if (!branchesOutput) return;

  const branches = branchesOutput.split('\n')
    .map(b => b.trim())
    .filter(b => b.startsWith('origin/') && !b.includes('origin/HEAD') && !b.includes('origin/main'));

  let mergedCount = 0;

  for (const branch of branches) {
    console.log(`Attempting to merge ${branch}...`);
    const shortBranch = branch.replace('origin/', '');
    
    const mergeResult = run(`git merge ${branch} --no-edit`);
    if (mergeResult) {
      console.log(`Successfully merged ${branch}. Deleting remote...`);
      run(`git push origin --delete ${shortBranch}`);
      mergedCount++;
    } else {
      console.log(`Merge conflict with ${branch}. Aborting...`);
      run('git merge --abort');
    }
  }

  if (mergedCount > 0) {
    console.log(`Pushing ${mergedCount} merged branches to origin main...`);
    run('git push origin main');
  } else {
    console.log('No branches merged.');
  }
}

mergeAll();
