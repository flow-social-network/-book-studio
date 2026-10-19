import {getApp, getApps, initializeApp} from 'firebase/app';
import {getAuth} from 'firebase/auth';

const required = {
  apiKey: process.env.NEXT_PUBLIC_FIREBASE_API_KEY,
  authDomain: process.env.NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN,
  projectId: process.env.NEXT_PUBLIC_FIREBASE_PROJECT_ID,
  appId: process.env.NEXT_PUBLIC_FIREBASE_APP_ID,
};

export function firebaseAuth() {
  const missing = Object.entries(required).filter(([, value]) => !value).map(([key]) => key);
  if (missing.length) throw new Error(`Firebase client configuration missing: ${missing.join(', ')}`);
  const app = getApps().length ? getApp() : initializeApp({
    apiKey: required.apiKey!,
    authDomain: required.authDomain!,
    projectId: required.projectId!,
    appId: required.appId!,
    ...(process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID ? {messagingSenderId: process.env.NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID} : {}),
  });
  return getAuth(app);
}
