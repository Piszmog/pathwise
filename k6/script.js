import http from 'k6/http';
import { check, sleep } from 'k6';

const targetVUs = parseInt(__ENV.TARGET_VUS, 10) || 2000;
const rampupDuration = __ENV.RAMPUP_DURATION || '10s';
const sustainedDuration = __ENV.SUSTAINED_DURATION || '40s';
const rampdownDuration = __ENV.RAMPDOWN_DURATION || '10s';

export const options = {
  discardResponseBodies: true,
  scenarios: {
    loadTest: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        {
          duration: rampupDuration,
          target: targetVUs,
        },
        {
          duration: sustainedDuration,
          target: targetVUs,
        },
        {
          duration: rampdownDuration,
          target: 0,
        }
      ],
      gracefulRampDown: '0s',
    },
  },
};

export default function() {
  const res = http.get('http://localhost:8080/signin');
  sleep(0.5);

  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}

