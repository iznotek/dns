import React, { Component } from 'react';
import {Redirect, Route, Switch, withRouter} from 'react-router-dom';
import {
    EuiHeader,
    EuiHeaderSection,
    EuiHeaderSectionItem,
    EuiHeaderSectionItemButton,
    EuiHeaderLogo,
    EuiPopover,
    EuiAvatar,
    EuiFlexGroup,
    EuiFlexItem,
    EuiSpacer,
    EuiLink,
    EuiText,
    EuiHeaderLinks,
    EuiHeaderLink,
    EuiHealth,
    EuiIcon
} from '@elastic/eui';

import Authentication from '../user';

import NotFound from './NotFound';
import Login from './Login';

const Records = () => <h2>Records</h2>;
const Users = () => <h2>Users</h2>;
const Roles = () => <h2>Roles</h2>;
const Profile = () => <h2>Profile</h2>;

class Base extends Component {
    constructor(props) {
        super(props);

        this.state = {
            userOpen: false,
            status: "success",
            loggedIn: Authentication.isAuthenticated()
        }
    }

    toggleUserMenuButtonClick = () => (this.state.loggedIn) ? this.setState({userOpen: !this.state.userOpen}) : "";

    onLogin = () => this.setState({loggedIn: Authentication.isAuthenticated()});
    onLogout = () => {
        Authentication.reset();
        this.setState({userOpen: !this.state.userOpen, loggedIn: false});
        this.forceUpdate();
    };

    render() {
        return (
            <div>
                <EuiHeader>
                    <EuiHeaderSection grow={true}>
                        <EuiHeaderSectionItem border="right">
                            <EuiHeaderLogo iconType="indexManagementApp" href={(Authentication.isAuthenticated()) ? "#/records" : "#"} aria-label="Go to home page">DNS Management</EuiHeaderLogo>
                        </EuiHeaderSectionItem>
                        <EuiHeaderLinks>
                            { Authentication.isAuthenticated() && <EuiHeaderLink href="#/records" isActive>Records</EuiHeaderLink> }
                            { Authentication.getUser().role === "admin" && <EuiHeaderLink href="#/users">Users</EuiHeaderLink> }
                            { Authentication.getUser().role === "admin" && <EuiHeaderLink href="#/roles">Roles</EuiHeaderLink> }
                        </EuiHeaderLinks>
                    </EuiHeaderSection>

                    <EuiHeaderSection side="right">
                        <EuiHeaderSectionItem>
                            <EuiPopover
                                id="headerUserMenu"
                                ownFocus
                                button={
                                    <EuiHeaderSectionItemButton
                                        aria-controls="headerUserMenu"
                                        aria-expanded={this.state.userOpen}
                                        aria-haspopup="true"
                                        aria-label="Account menu"
                                        onClick={this.toggleUserMenuButtonClick.bind(this)}>
                                        { !this.state.loggedIn && <EuiIcon type="lock" size="m"/> }
                                        { this.state.loggedIn && <EuiAvatar name={Authentication.getUser().name} size="s" />}
                                    </EuiHeaderSectionItemButton>
                                }
                                isOpen={this.state.userOpen}
                                anchorPosition="downRight"
                                closePopover={this.toggleUserMenuButtonClick.bind(this)}
                                panelPaddingSize="none">
                                <div style={{ width: 320 }}>
                                    <EuiFlexGroup
                                        gutterSize="m"
                                        className="euiHeaderProfile"
                                        responsive={false}>
                                        <EuiFlexItem grow={false}>
                                            <EuiAvatar name={Authentication.getUser().name} size="xl" />
                                        </EuiFlexItem>

                                        <EuiFlexItem>
                                            <EuiText>
                                                <p>{Authentication.getUser().name}</p>
                                            </EuiText>

                                            <EuiSpacer size="m" />

                                            <EuiFlexGroup>
                                                <EuiFlexItem>
                                                    <EuiFlexGroup justifyContent="spaceAround">
                                                        <EuiFlexItem grow={false}>
                                                            <EuiLink href="#/profile">Edit profile</EuiLink>
                                                        </EuiFlexItem>

                                                        <EuiFlexItem grow={false}>
                                                            <EuiLink onClick={this.onLogout.bind(this)}>Log out</EuiLink>
                                                        </EuiFlexItem>
                                                    </EuiFlexGroup>
                                                </EuiFlexItem>
                                            </EuiFlexGroup>
                                        </EuiFlexItem>
                                    </EuiFlexGroup>
                                </div>
                            </EuiPopover>
                        </EuiHeaderSectionItem>
                        <EuiHeaderSectionItem>
                            <EuiHeaderSectionItemButton>
                                <EuiHealth color={this.state.status}/>
                            </EuiHeaderSectionItemButton>
                        </EuiHeaderSectionItem>
                    </EuiHeaderSection>
                </EuiHeader>

                <Switch>
                    { !Authentication.isAuthenticated() && <Route exact path="/" render={(props) => <Login {...props} reload={this.forceUpdate.bind(this)} loginCb={this.onLogin.bind(this)}/>}/> }
                    { !Authentication.isAuthenticated() && <Redirect from="/" to="/"/>}

                    { Authentication.isAuthenticated() && <Redirect exact from="/" to="/records"/>}
                    { Authentication.isAuthenticated() && <Route path="/records" component={Records}/> }
                    { Authentication.isAuthenticated() && Authentication.getUser().role === "admin" && <Route path="/users" component={Users}/> }
                    { Authentication.isAuthenticated() && Authentication.getUser().role === "admin" && <Route path="/roles" component={Roles}/> }
                    { Authentication.isAuthenticated() && <Route path="/profile" component={Profile}/> }
                    <Route component={NotFound}/>
                </Switch>
            </div>
        )
    }
}

export default withRouter(Base);