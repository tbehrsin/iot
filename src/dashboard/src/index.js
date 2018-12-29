
import './error';
import React from 'react';
import { render } from 'react-dom';
import Application from './components/Application';

const reactRoot = document.getElementById('react');
render(<Application />, reactRoot);
