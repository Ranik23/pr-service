import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

const successCount = new Counter("success_count");
const failureCount = new Counter("failure_count");
const successRate = new Rate("success_rate");

export const options = {
  stages: [
    { duration: "5s", target: 10 }, 
    { duration: "1m", target: 10 }, 
  ],
  thresholds: {
    http_req_duration: ["p(95)<300"],
    http_req_failed: ["rate<0.001"],
    success_rate: ["rate>0.999"],
  },
};

const BASE_URL = "http://localhost:8080";

const TEAMS = [
  "backend", "frontend", "mobile", "payments", "infra",
  "data", "qa", "devops", "security", "ai"
];

export function setup() {
  console.log("Setting up 10 teams with 20 users each...");
  
  TEAMS.forEach(teamName => {
    const members = [];
    
    for (let i = 1; i <= 20; i++) {
      members.push({
        user_id: `user-${teamName}-${i}`,
        username: `User ${teamName}-${i}`,
        is_active: true
      });
    }
    
    const teamPayload = {
      team_name: teamName,
      members: members
    };

    const res = http.post(`${BASE_URL}/team/add`, JSON.stringify(teamPayload), {
      headers: { "Content-Type": "application/json" },
      timeout: "30s"
    });
    
    console.log(`Team ${teamName}: ${res.status}`);
  });
  
  console.log("Setup completed - 10 teams with 200 total users");
  return { TEAMS };
}

function generateUniqueId() {
  return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}-${__VU}-${__ITER}`;
}

export default function (data) {
  const vuId = __VU;

  const teamIndex = (vuId - 1) % data.TEAMS.length;
  const teamName = data.TEAMS[teamIndex];
  
  const userNumber = Math.floor(Math.random() * 20) + 1;
  const userId = `user-${teamName}-${userNumber}`;
  
  const uniqueId = generateUniqueId();
  const prId = `pr-${uniqueId}`;
  const prName = `PR-${uniqueId}`;
  
  const prPayload = {
    pull_request_id: prId,
    pull_request_name: prName,
    author_id: userId,
  };

  const prRes = http.post(`${BASE_URL}/pullRequest/create`, JSON.stringify(prPayload), {
    headers: { "Content-Type": "application/json" },
    timeout: "10s"
  });

  const prSuccess = check(prRes, {
    "PR created successfully": (r) => r.status === 201,
    "PR creation within SLI": (r) => r.timings.duration < 300
  });

  if (!prSuccess) {
    failureCount.add(1);
    successRate.add(0);
    return;
  }


  const mergePayload = { 
    pull_request_id: prId 
  };

  const mergeRes = http.post(`${BASE_URL}/pullRequest/merge`, JSON.stringify(mergePayload), {
    headers: { "Content-Type": "application/json" },
    timeout: "10s"
  });

  const mergeSuccess = check(mergeRes, {
    "PR merged successfully": (r) => r.status === 200,
    "PR merge within SLI": (r) => r.timings.duration < 300
  });

  if (!mergeSuccess) {
    failureCount.add(1);
    successRate.add(0);
    return;
  }


  if (prSuccess && mergeSuccess) {
    successCount.add(1);
    successRate.add(1);
  }


  sleep(0.8 + Math.random() * 0.4);
}