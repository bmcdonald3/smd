# MIT License
#
# (C) Copyright [2023] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

import json
import logging

from box import Box
from tavern.util import exceptions

logger = logging.getLogger(__name__)

def get_id_of_scn_subscriber_url(response, subscriber_url):
    subscriber_url_id = None
    response_data = json.loads(response.text)

    # check subscriptions for matching subscriber and url
    for subscription in response_data['SubscriptionList']:
        id = subscription['ID']
        subscriber = subscription['Subscriber']
        url = subscription['Url']
        logger.debug("Subscription: ID=%s, Subscriber=%s, Url=%s", id, subscriber, url)
        subscription_sub_url = subscriber + url
        if subscription_sub_url == subscriber_url:
            logger.debug("Found matching subscriber_url: ID=%s", id)
            subscriber_url_id = id
            break

    return Box({"subscriber_url_id": subscriber_url_id})