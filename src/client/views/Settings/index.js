import React from 'react';

import {
  StyleSheet,
  Text,
  TouchableWithoutFeedback,
  View
} from 'react-native';
import { connect } from 'react-redux';
import PropTypes from 'prop-types';
import { Route, Switch, Link } from 'react-router-native';
import Auth from './auth';
import { constants } from '../../../../app.json';
import * as actions from '../../actions';
import * as selectors from '../../selectors';
import { index as styles } from './styles';

class Settings extends React.Component {

  renderMenu = (props) => {
    return (
      <View style={styles.container}>
        <View style={styles.menuGroup}>
          <Link to={{ pathname: '/settings/auth', state: { animated: false, title: 'Authentication' } }} style={styles.menuItem} activeOpacity={0.8} underlayColor={constants.backgroundColor}>
            <Text style={styles.menuItemText}>Authentication</Text>
          </Link>
        </View>
      </View>
    )
  }

  render() {
    const { device, url, authToken } = this.props;

    return (
      <View style={styles.container}>
        <Switch>
          <Route exact path="/settings/" render={this.renderMenu} />
          <Route exact path="/settings/auth" component={Auth} />
        </Switch>
      </View>
    );
  }
}

const mapDispatchToProps = (dispatch, props) => ({
  request: (key, options) => dispatch(actions.api.request(key, options)),
  ...props
});
const mapStateToProps = (state, props) => ({

  ...props
});
export default connect(mapStateToProps, mapDispatchToProps)(Settings);
