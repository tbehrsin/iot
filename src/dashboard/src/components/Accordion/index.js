
import React from 'react';
import styles from './index.scss';

class Accordion extends React.Component {
  constructor() {
    super();
    this.state = {
      show: false
    };
  }

  onClick = () => {
    const { show } = this.state;
    this.setState({ show: !show });
  }

  render() {
    const { className, children } = this.props;
    const { show } = this.state;

    return (
      <div className={`${styles.container} ${className ? className : ''}`}>
        <div className={styles.header} onClick={this.onClick}>
          {React.Children.toArray(children)[0]}
          <div className={styles.spacer} />
          <i className={`fas fa-chevron-right ${show ? 'fa-rotate-90' : ''}`} />
        </div>
        {show && (
          <div className={styles.content}>
            {React.Children.toArray(children).slice(1)}
          </div>
        )}
      </div>
    )
  }
}

export default Accordion;
