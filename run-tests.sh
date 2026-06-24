kill $(lsof -t -i :3000) 2>/dev/null || true
cd frontend
npm run start &
sleep 5
npx playwright test
