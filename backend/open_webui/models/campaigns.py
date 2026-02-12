import logging
import time
from typing import Optional
import uuid

from sqlalchemy.orm import Session
from open_webui.internal.db import Base, get_db_context

from pydantic import BaseModel, ConfigDict
from sqlalchemy import BigInteger, Column, Text, JSON


log = logging.getLogger(__name__)

####################
# Campaign DB Schema
####################


class Campaign(Base):
    __tablename__ = "campaign"

    id = Column(Text, unique=True, primary_key=True)
    user_id = Column(Text)

    name = Column(Text)
    description = Column(Text)
    status = Column(Text)  # draft, active, paused, completed

    data = Column(JSON, nullable=True)
    meta = Column(JSON, nullable=True)

    created_at = Column(BigInteger)
    updated_at = Column(BigInteger)


class CampaignModel(BaseModel):
    id: str
    user_id: str

    name: str
    description: str
    status: str  # draft, active, paused, completed

    data: Optional[dict] = None
    meta: Optional[dict] = None

    created_at: int  # timestamp in epoch
    updated_at: int  # timestamp in epoch

    model_config = ConfigDict(from_attributes=True)


####################
# Forms
####################


class CampaignForm(BaseModel):
    name: str
    description: str
    status: Optional[str] = "draft"
    data: Optional[dict] = None
    meta: Optional[dict] = None


class CampaignUpdateForm(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    status: Optional[str] = None
    data: Optional[dict] = None
    meta: Optional[dict] = None


class CampaignResponse(CampaignModel):
    pass


####################
# Table
####################


class CampaignTable:
    def insert_new_campaign(
        self, user_id: str, form_data: CampaignForm, db: Optional[Session] = None
    ) -> Optional[CampaignModel]:
        with get_db_context(db) as db:
            campaign = CampaignModel(
                **{
                    **form_data.model_dump(exclude_none=True),
                    "id": str(uuid.uuid4()),
                    "user_id": user_id,
                    "status": form_data.status or "draft",
                    "created_at": int(time.time()),
                    "updated_at": int(time.time()),
                }
            )

            try:
                result = Campaign(**campaign.model_dump())
                db.add(result)
                db.commit()
                db.refresh(result)
                if result:
                    return CampaignModel.model_validate(result)
                else:
                    return None
            except Exception:
                return None

    def get_all_campaigns(
        self, db: Optional[Session] = None
    ) -> list[CampaignModel]:
        with get_db_context(db) as db:
            campaigns = (
                db.query(Campaign).order_by(Campaign.updated_at.desc()).all()
            )
            return [CampaignModel.model_validate(c) for c in campaigns]

    def get_campaigns_by_user_id(
        self, user_id: str, db: Optional[Session] = None
    ) -> list[CampaignModel]:
        with get_db_context(db) as db:
            campaigns = (
                db.query(Campaign)
                .filter_by(user_id=user_id)
                .order_by(Campaign.updated_at.desc())
                .all()
            )
            return [CampaignModel.model_validate(c) for c in campaigns]

    def get_campaign_by_id(
        self, id: str, db: Optional[Session] = None
    ) -> Optional[CampaignModel]:
        try:
            with get_db_context(db) as db:
                campaign = db.query(Campaign).filter_by(id=id).first()
                return CampaignModel.model_validate(campaign) if campaign else None
        except Exception:
            return None

    def update_campaign_by_id(
        self,
        id: str,
        form_data: CampaignUpdateForm,
        db: Optional[Session] = None,
    ) -> Optional[CampaignModel]:
        try:
            with get_db_context(db) as db:
                db.query(Campaign).filter_by(id=id).update(
                    {
                        **form_data.model_dump(exclude_none=True),
                        "updated_at": int(time.time()),
                    }
                )
                db.commit()
                return self.get_campaign_by_id(id=id, db=db)
        except Exception as e:
            log.exception(e)
            return None

    def delete_campaign_by_id(
        self, id: str, db: Optional[Session] = None
    ) -> bool:
        try:
            with get_db_context(db) as db:
                db.query(Campaign).filter_by(id=id).delete()
                db.commit()
                return True
        except Exception:
            return False


Campaigns = CampaignTable()