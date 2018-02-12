# Copyright 2017 Workiva
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import unittest

from frugal.util import make_hashable


class TestMakeHashable(unittest.TestCase):
    def test_make_hashable_basic(self):
        s = 'asdfadsf'
        self.assertEqual(s, make_hashable(s))

    def test_make_hashable_complex(self):
        data = {
            1: [[{1, 2}, {3, 4}], [{5, 6}, {7, 8}]],
            2: [[{10, 11}, {12, 13}], [{14, 15}, {16, 17}]],
        }
        expected = (
            (
                1,
                (
                    (frozenset([1, 2]), frozenset([3, 4])),
                    (frozenset([5, 6]), frozenset([7, 8])),
                ),
            ),
            (
                2,
                (
                    (frozenset([10, 11]), frozenset([12, 13])),
                    (frozenset([14, 15]), frozenset([16, 17])),
                )
            ),
        )
        # make sure that doesn't throw an error
        hash(expected)
        self.assertEqual(expected, make_hashable(data))


