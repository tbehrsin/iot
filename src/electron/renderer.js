
import React from 'react';
import { render } from 'react-dom';
import './renderer.scss';


const Application = () => (
  <div>
    Hello World
  </div>
);

render(<Application/>, document.getElementById('react-root'));
