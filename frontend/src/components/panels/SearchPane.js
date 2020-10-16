/*
 * This file is part of caronte (https://github.com/eciavatta/caronte).
 * Copyright (c) 2020 Emiliano Ciavatta.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

import React, {Component} from "react";
import Table from "react-bootstrap/Table";
import backend from "../../backend";
import dispatcher from "../../dispatcher";
import {createCurlCommand, dateTimeToTime, durationBetween} from "../../utils";
import ButtonField from "../fields/ButtonField";
import CheckField from "../fields/CheckField";
import InputField from "../fields/InputField";
import TagField from "../fields/TagField";
import TextField from "../fields/TextField";
import LinkPopover from "../objects/LinkPopover";
import "./common.scss";
import "./SearchPane.scss";

const _ = require("lodash");

class SearchPane extends Component {

    searchOptions = {
        "text_search": {
            "terms": null,
            "excluded_terms": null,
            "exact_phrase": "",
            "case_sensitive": false
        },
        "regex_search": {
            "pattern": "",
            "not_pattern": "",
            "case_insensitive": false,
            "multi_line": false,
            "ignore_whitespaces": false,
            "dot_character": false
        },
        "timeout": 10
    };

    state = {
        searches: [],
        currentSearchOptions: this.searchOptions,
    };

    componentDidMount() {
        this.reset();
        this.loadSearches();

        dispatcher.register("notifications", this.handleNotification);
        document.title = "caronte:~/searches$";
    }

    componentWillUnmount() {
        dispatcher.unregister(this.handleNotification);
    }

    loadSearches = () => {
        backend.get("/api/searches")
            .then((res) => this.setState({searches: res.json, searchesStatusCode: res.status}))
            .catch((res) => this.setState({searchesStatusCode: res.status, searchesResponse: JSON.stringify(res.json)}));
    };

    performSearch = () => {
        const options = this.state.currentSearchOptions;
        this.setState({loading: true});
        if (this.validateSearch(options)) {
            backend.post("/api/searches/perform", options).then((res) => {
                this.reset();
                this.setState({searchStatusCode: res.status, loading: false});
                this.loadSearches();
                this.viewSearch(res.json.id);
            }).catch((res) => {
                this.setState({
                    searchStatusCode: res.status, searchResponse: JSON.stringify(res.json),
                    loading: false
                });
            });
        }
    };

    reset = () => {
        this.setState({
            currentSearchOptions: _.cloneDeep(this.searchOptions),
            exactPhraseError: null,
            patternError: null,
            notPatternError: null,
            searchStatusCode: null,
            searchesStatusCode: null,
            searchResponse: null,
            searchesResponse: null
        });
    };

    validateSearch = (options) => {
        let valid = true;
        if (options["text_search"]["exact_phrase"] && options["text_search"]["exact_phrase"].length < 3) {
            this.setState({exactPhraseError: "text_search.exact_phrase.length < 3"});
            valid = false;
        }
        if (options["regex_search"].pattern && options["regex_search"].pattern.length < 3) {
            this.setState({patternError: "regex_search.pattern.length < 3"});
            valid = false;
        }
        if (options["regex_search"]["not_pattern"] && options["regex_search"]["not_pattern"].length < 3) {
            this.setState({exactPhraseError: "regex_search.not_pattern.length < 3"});
            valid = false;
        }

        return valid;
    };

    updateParam = (callback) => {
        callback(this.state.currentSearchOptions);
        this.setState({currentSearchOptions: this.state.currentSearchOptions});
    };

    extractPattern = (options) => {
        let pattern = "";
        if (_.isEqual(options.regex_search, this.searchOptions.regex_search)) { // is text search
            if (options["text_search"]["exact_phrase"]) {
                pattern += `"${options["text_search"]["exact_phrase"]}"`;
            } else {
                pattern += options["text_search"].terms.join(" ");
                if (options["text_search"]["excluded_terms"]) {
                    pattern += " -" + options["text_search"]["excluded_terms"].join(" -");
                }
            }
            options["text_search"]["case_sensitive"] && (pattern += "/s");
        } else { // is regex search
            if (options["regex_search"].pattern) {
                pattern += "/" + options["regex_search"].pattern + "/";
            } else {
                pattern += "!/" + options["regex_search"]["not_pattern"] + "/";
            }
            options["regex_search"]["case_insensitive"] && (pattern += "i");
            options["regex_search"]["multi_line"] && (pattern += "m");
            options["regex_search"]["ignore_whitespaces"] && (pattern += "x");
            options["regex_search"]["dot_character"] && (pattern += "s");
        }

        return pattern;
    };

    viewSearch = (searchId) => {
        dispatcher.dispatch("connections_filters", {"performed_search": searchId});
    };

    handleNotification = (payload) => {
        if (payload.event === "searches.new") {
            this.loadSearches();
        }
    };

    render() {
        const options = this.state.currentSearchOptions;

        let searches = this.state.searches.map((s) =>
            <tr key={s.id} className="row-small row-clickable">
                <td>{s.id.substring(0, 8)}</td>
                <td>{this.extractPattern(s["search_options"])}</td>
                <td>{s["affected_connections_count"]}</td>
                <td>{dateTimeToTime(s["started_at"])}</td>
                <td>{durationBetween(s["started_at"], s["finished_at"])}</td>
                <td><ButtonField name="view" variant="green" small onClick={() => this.viewSearch(s.id)}/></td>
            </tr>
        );

        const textOptionsModified = !_.isEqual(this.searchOptions.text_search, options.text_search);
        const regexOptionsModified = !_.isEqual(this.searchOptions.regex_search, options.regex_search);

        const curlCommand = createCurlCommand("/searches/perform", "POST", options);

        return (
            <div className="pane-container search-pane">
                <div className="pane-section searches-list">
                    <div className="section-header">
                        <span className="api-request">GET /api/searches</span>
                        {this.state.searchesStatusCode &&
                        <span className="api-response"><LinkPopover text={this.state.searchesStatusCode}
                                                                    content={this.state.searchesResponse}
                                                                    placement="left"/></span>}
                    </div>

                    <div className="section-content">
                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>id</th>
                                    <th>pattern</th>
                                    <th>occurrences</th>
                                    <th>started_at</th>
                                    <th>duration</th>
                                    <th>actions</th>
                                </tr>
                                </thead>
                                <tbody>
                                {searches}
                                </tbody>
                            </Table>
                        </div>
                    </div>
                </div>

                <div className="pane-section search-new">
                    <div className="section-header">
                        <span className="api-request">POST /api/searches/perform</span>
                        <span className="api-response"><LinkPopover text={this.state.searchStatusCode}
                                                                    content={this.state.searchResponse}
                                                                    placement="left"/></span>
                    </div>

                    <div className="section-content">
                        <span className="notes">
                        NOTE: it is recommended to use the rules for recurring themes. Give preference to textual search over that with regex.
                    </span>

                        <div className="content-row">
                            <div className="text-search">
                                <TagField tags={(options["text_search"].terms || []).map((t) => {
                                    return {name: t};
                                })}
                                          name="terms" min={3} inline allowNew={true}
                                          readonly={regexOptionsModified || options["text_search"]["exact_phrase"]}
                                          onChange={(tags) => this.updateParam((s) => s["text_search"].terms = tags.map((t) => t.name))}/>
                                <TagField tags={(options["text_search"]["excluded_terms"] || []).map((t) => {
                                    return {name: t};
                                })}
                                          name="excluded_terms" min={3} inline allowNew={true}
                                          readonly={regexOptionsModified || options["text_search"]["exact_phrase"]}
                                          onChange={(tags) => this.updateParam((s) => s["text_search"]["excluded_terms"] = tags.map((t) => t.name))}/>

                                <span className="exclusive-separator">or</span>

                                <InputField name="exact_phrase" value={options["text_search"]["exact_phrase"]} inline
                                            error={this.state.exactPhraseError}
                                            onChange={(v) => this.updateParam((s) => s["text_search"]["exact_phrase"] = v)}
                                            readonly={regexOptionsModified || (Array.isArray(options["text_search"].terms) && options["text_search"].terms.length > 0)}/>

                                <CheckField checked={options["text_search"]["case_sensitive"]} name="case_sensitive"
                                            readonly={regexOptionsModified} small
                                            onChange={(v) => this.updateParam((s) => s["text_search"]["case_sensitive"] = v)}/>
                            </div>

                            <div className="separator">
                                <span>or</span>
                            </div>

                            <div className="regex-search">
                                <InputField name="pattern" value={options["regex_search"].pattern} inline
                                            error={this.state.patternError}
                                            readonly={textOptionsModified || options["regex_search"]["not_pattern"]}
                                            onChange={(v) => this.updateParam((s) => s["regex_search"].pattern = v)}/>
                                <span className="exclusive-separator">or</span>
                                <InputField name="not_pattern" value={options["regex_search"]["not_pattern"]} inline
                                            error={this.state.notPatternError}
                                            readonly={textOptionsModified || options["regex_search"].pattern}
                                            onChange={(v) => this.updateParam((s) => s["regex_search"]["not_pattern"] = v)}/>

                                <div className="checkbox-line">
                                    <CheckField checked={options["regex_search"]["case_insensitive"]}
                                                name="case_insensitive"
                                                readonly={textOptionsModified} small
                                                onChange={(v) => this.updateParam((s) => s["regex_search"]["case_insensitive"] = v)}/>
                                    <CheckField checked={options["regex_search"]["multi_line"]} name="multi_line"
                                                readonly={textOptionsModified} small
                                                onChange={(v) => this.updateParam((s) => s["regex_search"]["multi_line"] = v)}/>
                                    <CheckField checked={options["regex_search"]["ignore_whitespaces"]}
                                                name="ignore_whitespaces"
                                                readonly={textOptionsModified} small
                                                onChange={(v) => this.updateParam((s) => s["regex_search"]["ignore_whitespaces"] = v)}/>
                                    <CheckField checked={options["regex_search"]["dot_character"]} name="dot_character"
                                                readonly={textOptionsModified} small
                                                onChange={(v) => this.updateParam((s) => s["regex_search"]["dot_character"] = v)}/>
                                </div>
                            </div>
                        </div>

                        <TextField value={curlCommand} rows={3} readonly small={true}/>
                    </div>

                    <div className="section-footer">
                        <ButtonField variant="red" name="cancel" bordered disabled={this.state.loading}
                                     onClick={this.reset}/>
                        <ButtonField variant="green" name="perform_search" bordered
                                     disabled={this.state.loading} onClick={this.performSearch}/>
                    </div>
                </div>
            </div>
        );
    }

}

export default SearchPane;
