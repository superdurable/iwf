/*
 * Copyright (c) 2022-2026 Super Durable, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package io.iworkflow.core;

import feign.FeignException;
import io.iworkflow.core.exceptions.LongPollTimeoutException;
import io.iworkflow.core.exceptions.NoRunningWorkflowException;
import io.iworkflow.core.exceptions.WorkflowAlreadyStartedException;
import io.iworkflow.gen.models.EncodedObject;
import io.iworkflow.gen.models.ErrorResponse;
import io.iworkflow.gen.models.ErrorSubStatus;

import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;
import java.util.Optional;

public abstract class IwfHttpException extends RuntimeException {

    private final int statusCode;
    private ErrorResponse errorResponse;

    public IwfHttpException(final ObjectEncoder objectEncoder, final FeignException.FeignClientException exception) {
        super(exception);
        statusCode = exception.status();
        String decodeErrorMessage = "";
        final Optional<ByteBuffer> respBody = exception.responseBody();
        if (respBody.isPresent()) {
            String data = StandardCharsets.UTF_8.decode(respBody.get()).toString();
            try {
                errorResponse = objectEncoder.decode(new EncodedObject().data(data), ErrorResponse.class);
                return;
            } catch (Exception e) {
                decodeErrorMessage = e.getMessage();
            }
        }
        errorResponse = new ErrorResponse()
                .detail("empty or unable to decode to ErrorResponse:" + decodeErrorMessage)
                .subStatus(ErrorSubStatus.UNCATEGORIZED_SUB_STATUS);
    }

    protected IwfHttpException(final IwfHttpException exception) {
        statusCode = exception.getStatusCode();
        errorResponse = exception.getErrorResponse();
    }

    public IwfHttpException() {
        statusCode = 500;
    }

    public String getErrorDetails() {
        return errorResponse.getDetail();
    }

    public int getStatusCode() {
        return statusCode;
    }

    public ErrorSubStatus getErrorSubStatus() {
        return errorResponse.getSubStatus();
    }

    public ErrorResponse getErrorResponse() {
        return errorResponse;
    }

    public static IwfHttpException fromFeignException(final ObjectEncoder objectEncoder, final FeignException.FeignClientException exception) {
        if (exception.status() >= 400 && exception.status() < 500) {
            final ClientSideException clientSideException = new ClientSideException(objectEncoder, exception);

            switch (clientSideException.getErrorSubStatus()) {
                case LONG_POLL_TIME_OUT_SUB_STATUS:
                    return new LongPollTimeoutException(clientSideException);
                case WORKFLOW_ALREADY_STARTED_SUB_STATUS:
                    return new WorkflowAlreadyStartedException(clientSideException);
                case WORKFLOW_NOT_EXISTS_SUB_STATUS:
                    return new NoRunningWorkflowException(clientSideException);
                default:
                    return clientSideException;
            }
        } else {
            return new ServerSideException(objectEncoder, exception);
        }
    }
}
