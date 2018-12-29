
import React from 'react';
import { TransitionGroup, CSSTransition } from 'react-transition-group';

import styles from './index.scss';

class Menu extends React.Component {
  constructor() {
    super();
    this.state = {
      show: false
    };
  }

  show = () => {
    this.setState({ show: true });
  }

  hide = () => {
    this.setState({ show: false });
  }

  toggle = () => {
    const { show } = this.state;
    this.setState({ show: !show });
  }

  render() {
    const { children, className } = this.props;
    const { show } = this.state;

    const candidates = React.Children.toArray(children).filter((child) => location.pathname.startsWith(child.props.to));
    candidates.sort((a, b) => {
      if (a.props.to < b.props.to) {
        return -1;
      }
      if (a.props.to > b.props.to) {
        return 1;
      }
      return 0;
    });
    const child = candidates.reverse()[0];

    return (
      <div className={styles.container}>
        {show && (
          <div className={styles.background} onClick={this.hide}/>
        )}
        <button className={styles.button} onClick={this.toggle}>
          <span>{child ? child.props.children : 'Menu'}</span>
          <i className="fas fa-chevron-down" />
        </button>
        <div className={styles.mask} onClick={this.hide}>
          <TransitionGroup>
            {show && (
              <CSSTransition timeout={400} in classNames="menu">
                <ul className={`${styles.dropDown} ${className ? className : ''}`}>
                  {React.Children.map(children, (child, i) => <li key={i}>{child}</li>)}
                </ul>
              </CSSTransition>
            )}
          </TransitionGroup>
        </div>
      </div>
    );
  }
}

export default Menu;
