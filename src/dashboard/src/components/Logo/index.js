import React from 'react';
import styles from './index.scss';

class Logo extends React.Component {
  render() {
    return (
      <span className={styles.container}>
        {'Behrsin '}
        <span className={styles.iot}>{'IoT'}</span>
      </span>
    );
  }
}

export default Logo;
