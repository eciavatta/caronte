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

const log = {
    debug: (...obj) => isDevelopment() && console.info(...obj),
    info: (...obj) => console.info(...obj),
    warn: (...obj) => console.warn(...obj),
    error: (...obj) => console.error(obj)
};

function isDevelopment() {
    return !process.env.NODE_ENV || process.env.NODE_ENV === "development";
}

export default log;
