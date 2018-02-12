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

import six


def make_hashable(thing):
    """
    Creates a representation of some input data with immutable collections,
    so the output is hashable.

    :param thing: The data to create a hashable representation of.
    """
    if isinstance(thing, list):
        return tuple([make_hashable(elem) for elem in thing])
    elif isinstance(thing, set):
        return frozenset([make_hashable(elem) for elem in thing])
    elif isinstance(thing, dict):
        new_dict = {make_hashable(k): make_hashable(v)
                    for k, v in six.iteritems(thing)}
        return tuple(six.iteritems(new_dict))
    else:
        return thing
