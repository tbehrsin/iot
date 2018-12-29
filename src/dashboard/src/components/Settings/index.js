
import React from 'react';
import { connect } from 'react-redux';
import { Router, Route, Switch, Redirect, Link } from 'react-router-dom';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';
import { store, persistor } from '../../store';
import { history } from '../../router';

import * as actions from '../../actions';
import * as selectors from '../../selectors';
import { ws } from '../../services';

import Page from '../Page';
import Button from '../Button';
import DeveloperModeSettings from './developer-mode';

import styles from './index.scss';

const categories = [
  {
    path: '/dashboard/settings/developer',
    title: 'Developer Mode',
    component: DeveloperModeSettings
  }
]

class Settings extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
    };
  }

  componentDidMount() {
    const { device, request, match } = this.props;


  }

  render() {
    const { location } = this.props;

    return (
      <Page.FullScreen className={styles.container}>
        <div className={styles.columns}>
          <ul className={styles.links}>
            {categories.map(category => (
              <li key={category.path}><Link to={category.path} className={location.pathname === category.path ? styles.activeLink : ''}>{category.title}</Link></li>
            ))}
          </ul>
          <div>
            <Switch>
              {categories.map(category => (
                <Route exact path={category.path} key={category.path} component={category.component} />
              ))}
              <Redirect path="/dashboard/settings" to={categories[0].path} />
            </Switch>
          </div>
        </div>
      </Page.FullScreen>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  setIn: (key, path, value) => dispatch(actions.api.setIn(key, path, value)),
  ...props
});
const mapStateToProps = (state, props) => ({
  developerMode: selectors.api.body(state)('developer-mode'),
  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Settings);
