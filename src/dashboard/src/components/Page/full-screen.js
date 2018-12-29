
import React from 'react';
import { connect } from 'react-redux';
import { Link } from 'react-router-dom';

import { API_STATE_REQUESTING, API_STATE_ERROR } from '../../constants';
import * as actions from '../../actions';
import * as selectors from '../../selectors';

import Menu from '../Menu';

import styles from './full-screen.scss';

class FullScreenPage extends React.Component {
  render() {
    const { header, children, className, requestStateAll, connectionState } = this.props;

    return (
      <div className={styles.container}>
        <div className={styles.header}>
          <Link to="/dashboard/" className={styles.logo}>Behrsin <strong>IoT</strong></Link>
          <div className={styles.spacer} />
          {requestStateAll === API_STATE_REQUESTING && (
            <i className={`${styles.apiRequesting} fas fa-sync-alt fa-spin`} />
          )}
          {requestStateAll === API_STATE_ERROR && (
            <i className={`${styles.apiError} fas fa-exclamation-triangle`} />
          )}
          <div className={`${styles.connectionState} ${styles[connectionState]}`} />
          <Menu className={styles.menu}>
            <Link to="/dashboard/">Dashboard</Link>
            <Link to="/dashboard/settings/">Settings</Link>
            <Link to="/">Home</Link>
          </Menu>
        </div>
        <div className={`${styles.content} ${className}`}>
          {children}
        </div>
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
  connectionState: selectors.api.connectionState(state),
  requestStateAll: selectors.api.requestStateAll(state),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(FullScreenPage);
