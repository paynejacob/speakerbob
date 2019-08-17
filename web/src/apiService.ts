import axios from 'axios';
import { ValidateStatus } from '@/utils/auth';

export default axios.create({
  baseURL: `${window.location.protocol}//${window.location.hostname}:${window.location.port}/api/`,
  headers: {
    post: {
      'Content-Type': 'application/json',
    },
    put: {
      'Content-Type': 'application/json',
    },
  },
  validateStatus: ValidateStatus,
  withCredentials: true,
});
