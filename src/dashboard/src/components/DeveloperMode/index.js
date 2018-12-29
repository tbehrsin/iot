
import React from 'react';
import Accordion from '../Accordion';
import Button from '../Button';
import styles from './index.scss';

class DeveloperMode extends React.Component {

  onClick

  render() {
    const { device } = this.props;

    return (
      <Accordion className={styles.container}>
        <div>Developer Mode</div>
        <pre className={styles.shell} ref="code">
          {'$>'} <span className={styles.selectable}>{'npm install -g '}<strong>{'@behrsin/iot-cli'}</strong>{'\n'}</span>
          {'$>'} <span className={styles.selectable}><b>{'iot login 192.168.0.20\n'}</b></span>
          {'$>'} <span className={styles.selectable}>iot create my-app{'\n'}</span>
          {'$>'} <span className={styles.selectable}>{'cd my-app\n'}</span>
          {'$>'} <span className={styles.selectable}>iot serve <strong>{device.id}</strong>{'\n'}</span>
          <Button className={styles.copy} onClick={this.onClick}><i className="fas fa-paste" /></Button>
        </pre>
      </Accordion>
    )
  }
}

export default DeveloperMode;
