import React, { FC } from 'react';
import { Form, Stack } from 'react-bootstrap';

import type * as Type from '@/common/interface';

interface Props {
  type: 'radio' | 'checkbox';
  fieldName: string;
  onChange?: (fd: Type.FormDataType) => void;
  enumValues: (string | boolean | number)[];
  enumNames: string[];
  formData: Type.FormDataType;
  readOnly?: boolean;
}
const Index: FC<Props> = ({
  type = 'radio',
  fieldName,
  onChange,
  enumValues,
  enumNames,
  formData,
  readOnly = false,
}) => {
  const fieldObject = formData[fieldName];
  const handleCheck = (
    evt: React.ChangeEvent<HTMLInputElement>,
    index: number,
  ) => {
    const { name, checked } = evt.currentTarget;
    enumValues[index] = checked;

    const state = {
      ...formData,
      [name]: {
        ...formData[name],
        value: enumValues,
        isInvalid: false,
      },
    };
    if (typeof onChange === 'function') {
      console.log('fieldName', fieldName, enumValues);
      onChange(state);
    }
  };
  return (
    <Stack direction="horizontal">
      {enumValues?.map((item, index) => {
        return (
          <Form.Check
            key={String(item)}
            inline
            type={type}
            name={fieldName}
            id={`${fieldName}-${enumNames?.[index]}`}
            label={enumNames?.[index]}
            checked={fieldObject?.value?.[index] || false}
            feedback={fieldObject?.errorMsg}
            feedbackType="invalid"
            isInvalid={fieldObject?.isInvalid}
            disabled={readOnly}
            onChange={(evt) => handleCheck(evt, index)}
          />
        );
      })}
    </Stack>
  );
};

export default Index;
