
import React from 'react';
import { connect } from 'react-redux';
import { Link } from 'react-router-dom';

import * as actions from '../../actions';
import * as selectors from '../../selectors';

import Menu from '../Menu';
import FullScreen from './full-screen';

import styles from './index.scss';

class Page extends React.Component {
  componentDidMount() {
    this.onScroll();
    this.onResize();
    window.addEventListener('load', this.onResize, false);
    window.addEventListener('resize', this.onResize, false);

    const { header } = this.refs;
    for (const img of [...header.querySelectorAll('img')]) {
      img.addEventListener('load', this.onResize, false);
    }
  }

  componentWillUnmount() {
    window.removeEventListener('load', this.onResize, false);
    window.removeEventListener('resize', this.onResize, false);
  }

  onResize = () => {
    const { container, header } = this.refs;
    const { height } = header.getBoundingClientRect();

    container.style.paddingTop = `${height | 0}px`;
  };

  onScroll = () => {
    const { container, header, headerContent } = this.refs;

    const scrollTop = container.scrollTop;

    header.style.position = 'fixed';
    header.style.top = 0;
    header.style.left = 0;
    header.style.right = 0;
    headerContent.style.marginTop = `-${scrollTop | 0}px`;


  };

  render() {
    const { header, children, className, hasAppToken } = this.props;

    return (
      <div ref="container" className={styles.container} onScroll={this.onScroll}>
        <div ref="header" className={styles.header}>
          <div className={styles.navigation}>
            <div className={styles.logo}>Behrsin <strong>IoT</strong></div>
            <div className={styles.spacer} />

            <Menu>
              <Link to="/">Home</Link>
              {hasAppToken && <Link to="/dashboard">Dashboard</Link>}
            </Menu>
          </div>

          <div ref="headerContent" className={styles.headerContent}>
            {header && header()}
          </div>
        </div>
        <div className={`${styles.content} ${className}`}>
          {children}
        </div>
        <div className={styles.spacer} />
        <footer className={styles.footer}>
          &copy; {new Date().getFullYear()} Behrsin Ltd
        </footer>
      </div>
    )
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  ...props
});
const mapStateToProps = (state, props) => ({
  hasAppToken: selectors.auth.hasAppToken(state),
  ...props
});
export default Object.assign(connect(mapStateToProps, mapDispatchToProps)(Page), { FullScreen });
