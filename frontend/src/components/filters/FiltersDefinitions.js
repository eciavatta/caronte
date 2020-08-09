import {
    cleanNumber,
    timestampToTime,
    timeToTimestamp,
    validate24HourTime,
    validateIpAddress,
    validateMin,
    validatePort
} from "../../utils";
import StringConnectionsFilter from "./StringConnectionsFilter";
import React from "react";
import RulesConnectionsFilter from "./RulesConnectionsFilter";
import BooleanConnectionsFilter from "./BooleanConnectionsFilter";


export const filtersNames = ["service_port", "matched_rules", "client_address", "client_port",
    "min_duration", "max_duration", "min_bytes", "max_bytes", "started_after",
    "started_before", "closed_after", "closed_before", "marked", "hidden"];

export const filtersDefinitions = {
    service_port: <StringConnectionsFilter filterName="service_port"
                                           defaultFilterValue="all_ports"
                                           replaceFunc={cleanNumber}
                                           validateFunc={validatePort}/>,
    matched_rules: <RulesConnectionsFilter />,
    client_address: <StringConnectionsFilter filterName="client_address"
                                             defaultFilterValue="all_addresses"
                                             validateFunc={validateIpAddress} />,
    client_port: <StringConnectionsFilter filterName="client_port"
                                          defaultFilterValue="all_ports"
                                          replaceFunc={cleanNumber}
                                          validateFunc={validatePort}/>,
    min_duration: <StringConnectionsFilter filterName="min_duration"
                                           defaultFilterValue="0"
                                           replaceFunc={cleanNumber}
                                           validateFunc={validateMin(0)}/>,
    max_duration: <StringConnectionsFilter filterName="max_duration"
                                           defaultFilterValue="∞"
                                           replaceFunc={cleanNumber} />,
    min_bytes: <StringConnectionsFilter filterName="min_bytes"
                                        defaultFilterValue="0"
                                        replaceFunc={cleanNumber}
                                        validateFunc={validateMin(0)} />,
    max_bytes: <StringConnectionsFilter filterName="max_bytes"
                                        defaultFilterValue="∞"
                                        replaceFunc={cleanNumber} />,
    started_after: <StringConnectionsFilter filterName="started_after"
                                            defaultFilterValue="00:00:00"
                                            validateFunc={validate24HourTime}
                                            encodeFunc={timeToTimestamp}
                                            decodeFunc={timestampToTime} />,
    started_before: <StringConnectionsFilter filterName="started_before"
                                             defaultFilterValue="00:00:00"
                                             validateFunc={validate24HourTime}
                                             encodeFunc={timeToTimestamp}
                                             decodeFunc={timestampToTime} />,
    closed_after: <StringConnectionsFilter filterName="closed_after"
                                           defaultFilterValue="00:00:00"
                                           validateFunc={validate24HourTime}
                                           encodeFunc={timeToTimestamp}
                                           decodeFunc={timestampToTime} />,
    closed_before: <StringConnectionsFilter filterName="closed_before"
                                            defaultFilterValue="00:00:00"
                                            validateFunc={validate24HourTime}
                                            encodeFunc={timeToTimestamp}
                                            decodeFunc={timestampToTime} />,
    marked: <BooleanConnectionsFilter filterName={"marked"} />,
    hidden: <BooleanConnectionsFilter filterName={"hidden"} />
};
