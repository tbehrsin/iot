
import * as api from './api';
import * as auth from './auth';

export default [
  ...Object.values(api),
  ...Object.values(auth)
];
