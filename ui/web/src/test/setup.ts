import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

// Cleanup after each test
afterEach(() => {
  cleanup();
});

// Mock WebSocket
global.WebSocket = class WebSocket {
  CONNECTING = 0;
  OPEN = 1;
  CLOSING = 2;
  CLOSED = 3;
  readyState = this.OPEN;

  constructor(public url: string) {}

  send = () => {};
  close = () => {};
  addEventListener = () => {};
  removeEventListener = () => {};
  dispatchEvent = () => true;
} as any;
