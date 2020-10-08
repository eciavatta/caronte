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

const validation = {
    isValidColor: (color) => /^#(?:[0-9a-fA-F]{3}){1,2}$/.test(color),
    isValidPort: (port, required) => parseInt(port, 10) > (required ? 0 : -1) && parseInt(port, 10) <= 65565,
    isValidAddress: (address) => true // TODO
};

export default validation;
